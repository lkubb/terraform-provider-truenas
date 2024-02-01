package truenas

import (
	"context"
	"errors"
	api "github.com/dariusbakunas/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"regexp"
	"strconv"
)

func resourceTrueNASUser() *schema.Resource {
	return &schema.Resource{
		Description:   "In TrueNAS, user accounts allow flexibility for accessing shared data. A common practice is to create users and assign them to groups. This allows for efficient permissions tuning for large numbers of users.",
		CreateContext: resourceTrueNASUserCreate,
		ReadContext:   resourceTrueNASUserRead,
		UpdateContext: resourceTrueNASUserUpdate,
		DeleteContext: resourceTrueNASUserDelete,
		Schema: map[string]*schema.Schema{
			"user_id": &schema.Schema{
				Description: "User ID",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"uid": &schema.Schema{
				Description: "UID",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				// This is arbitrary. Sensible? @TODO
				ValidateFunc: validation.IntBetween(0, 65535),
			},
			"username": &schema.Schema{
				Description: "Username",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 16),
					validation.StringDoesNotContainAny(`	,:+&#%^()!@~*?<>= `),
					validation.StringMatch(regexp.MustCompile(`^[^\$-][^\$\s]*.$`), "Username must not begin with a hyphen and `$` can only be included as the last character."),
				),
			},
			"unixhash": &schema.Schema{
				Description: "Password hash (UNIX)",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"smbhash": &schema.Schema{
				Description: "Password hash (SMB)",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"home": &schema.Schema{
				Description: "This user's home directory. Must begin with /mnt, followed by a valid pool or disk name and be writable.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"shell": &schema.Schema{
				Description: "This user's shell executable, absolute path.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				// Validation is possible with an API call @TODO
			},
			"full_name": &schema.Schema{
				Description: "This user's full name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"builtin": &schema.Schema{
				Description: "Whether this is a builtin user.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"smb": &schema.Schema{
				Description: "Samba authentication: Set to allow user to authenticate to Samba shares.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"password": &schema.Schema{
				Description:  "This user's password.",
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringDoesNotContainAny("?"),
			},
			"password_disabled": &schema.Schema{
				Description:   "Password disabled: The account cannot use password-based logins for services.",
				Type:          schema.TypeBool,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"password"},
			},
			"locked": &schema.Schema{
				Description:   "Prevent the user from logging in or using password-based services until this option is unset. Locking an account is only possible when Disable Password is No and a Password has been created for the account.",
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"password_disabled"},
			},
			"sudo": &schema.Schema{
				Description: "Permit sudo usage by this user.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"sudo_nopasswd": &schema.Schema{
				Description: "Permit sudo usage without password authentication for this user.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"sudo_commands": &schema.Schema{
				Description: "List of permitted sudo commands for this user.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"microsoft_account": &schema.Schema{
				Description: "Microsoft account: Allow additional username authentication methods for Windows >=8 clients.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			// "attributes": &schema.Schema{
			// 	Description: "Attributes",
			// 	Type:        schema.TypeMap,
			// 	Computed:    true,
			// },
			"email": &schema.Schema{
				Description: "This user's email address.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"primary_group": &schema.Schema{
				Description: "Group ID of this user's primary group (resource ID, not GID). If set, needs to exist.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"primary_group_name": &schema.Schema{
				Description: "The name of this user's primary group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			// TrueNAS automatically adds the user
			// to `builtin_users` group if smb is true.
			// Therefore, groups is marked as computed.
			// This might be solved more elegantly with CustomizeDiff.
			"groups": &schema.Schema{
				Description: "List of group IDs this user is member in (resource IDs, not GIDs).",
				Type:        schema.TypeSet,
				Computed:    true,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"sshpubkey": &schema.Schema{
				Description: "This user's SSH pubkey.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"local": &schema.Schema{
				Description: "Whether this is a local user.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"id_type_both": &schema.Schema{
				Description: "Samba: ID_TYPE_BOTH",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func resourceTrueNASUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("error parsing group ID: %s", err)
	}

	resp, http, err := c.UserApi.GetUser(ctx, int32(id)).Execute()

	if err != nil {
		d.SetId("")
		if http.StatusCode == 404 {
			return nil
		}
		return diag.Errorf("error getting user: %s", err)
	}

	if resp.Uid != nil {
		d.Set("uid", *resp.Uid)
	}

	d.Set("username", resp.Username)

	if resp.Unixhash != nil {
		d.Set("unixhash", *resp.Unixhash)
	}

	if resp.Smbhash != nil {
		d.Set("smbhash", *resp.Smbhash)
	}

	if resp.Home != nil {
		d.Set("home", *resp.Home)
	}

	if resp.Shell != nil {
		d.Set("shell", *resp.Shell)
	}

	d.Set("full_name", resp.FullName)

	d.Set("builtin", *resp.Builtin)
	d.Set("smb", *resp.Smb)
	d.Set("password_disabled", *resp.PasswordDisabled)
	d.Set("locked", *resp.Locked)
	d.Set("sudo", *resp.Sudo)
	d.Set("sudo_nopasswd", *resp.SudoNopasswd)

	if err := d.Set("sudo_commands", flattenStringList(resp.SudoCommands)); err != nil {
		return diag.Errorf("error setting sudo_commands: %s", err)
	}

	d.Set("microsoft_account", *resp.MicrosoftAccount)

	// "attributes" has little value to TF afaict,
	// skip it for the time being. @TODO?

	// resp.Email.IsSet() was unreliable in my testing
	email := resp.Email.Get()
	if email != nil {
		d.Set("email", *email)
	}

	if resp.Group != nil {
		if resp.Group.Id != nil {
			d.Set("primary_group", *resp.Group.Id)
		}
		// This is helpful for a dumb heuristic to decide if
		// the user's primary group should be deleted if it
		// would be left empty after user deletion.
		if resp.Group.BsdgrpGroup != nil {
			d.Set("primary_group_name", *resp.Group.BsdgrpGroup)
		}
	}

	if err := d.Set("groups", flattenInt32List(resp.Groups)); err != nil {
		return diag.Errorf("error setting users: %s", err)
	}

	sshpubkey := resp.Sshpubkey.Get()
	if sshpubkey != nil {
		d.Set("sshpubkey", *sshpubkey)
	}

	d.Set("local", *resp.Local)
	d.Set("id_type_both", *resp.IdTypeBoth)

	d.Set("user_id", int(resp.Id))

	return diags
}

func resourceTrueNASUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)

	user, err := expandUser(d)

	if err != nil {
		return diag.Errorf("Error while parsing resource configuration: %s", err)
	}

	resp, _, err := c.UserApi.CreateUser(ctx).CreateUserParams(user).Execute()

	if err != nil {
		return diag.Errorf("error creating user: %s", err)
	}

	d.SetId(strconv.Itoa(int(resp)))

	return resourceTrueNASUserRead(ctx, d, m)
}

func resourceTrueNASUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)

	user := expandUserForDelete(d)

	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting TrueNAS user: %s", strconv.Itoa(id))

	_, err = c.UserApi.DeleteUser(ctx, int32(id)).DeleteUserParams(user).Execute()

	if err != nil {
		return diag.Errorf("error deleting user: %s", err)
	}

	log.Printf("[INFO] TrueNAS user (%s) deleted", strconv.Itoa(id))
	d.SetId("")

	return diags
}

func resourceTrueNASUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)

	user, err := expandUserForUpdate(d)
	// user := expandUserForUpdate(d)

	if err != nil {
		return diag.Errorf("Error while parsing resource configuration: %s", err)
	}

	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	_, _, err = c.UserApi.UpdateUser(ctx, int32(id)).UpdateUserParams(user).Execute()

	if err != nil {
		return diag.Errorf("error updating user: %s", err)
	}

	return resourceTrueNASUserRead(ctx, d, m)
}

func expandUser(d *schema.ResourceData) (api.CreateUserParams, error) {
	user := api.CreateUserParams{
		FullName: d.Get("full_name").(string),
		Username: d.Get("username").(string),
		// @TODO:
		GroupCreate: getBoolPtr(true),
		HomeMode:    getStringPtr("755"),
	}

	if uid, ok := d.GetOk("uid"); ok {
		user.Uid = getInt32Ptr(int32(uid.(int)))
	}

	if home, ok := d.GetOk("home"); ok {
		user.Home = getStringPtr(home.(string))
	}

	if shell, ok := d.GetOk("shell"); ok {
		user.Shell = getStringPtr(shell.(string))
	}

	smb := d.Get("smb")
	if smb != nil {
		user.Smb = getBoolPtr(smb.(bool))
	}

	if password, ok := d.GetOk("password"); ok {
		user.Password = getStringPtr(password.(string))
	}

	// disable password auth automatically if it was not set
	_, password_set := d.GetOk("password")
	user.PasswordDisabled = getBoolPtr(!password_set)

	locked := d.Get("locked")
	if locked != nil {
		// Making a user locked needs a password set.
		if locked.(bool) && *user.PasswordDisabled {
			return user, errors.New("Locked needs the user to have a password.")
		}
		user.Locked = getBoolPtr(locked.(bool))
	}

	sudo := d.Get("sudo")
	if sudo != nil {
		user.Sudo = getBoolPtr(sudo.(bool))
	}

	sudo_nopasswd := d.Get("sudo_nopasswd")
	if sudo_nopasswd != nil {
		user.SudoNopasswd = getBoolPtr(sudo_nopasswd.(bool))
	}

	if sudo_commands, ok := d.GetOk("sudo_commands"); ok {
		user.SudoCommands = expandStrings(sudo_commands.(*schema.Set).List())
	}

	microsoft_account := d.Get("microsoft_account")
	if microsoft_account != nil {
		user.MicrosoftAccount = getBoolPtr(microsoft_account.(bool))
	}

	if email, ok := d.GetOk("email"); ok {
		user.Email.Set(getStringPtr(email.(string)))
	}

	if primary_group, ok := d.GetOk("primary_group"); ok {
		user.Group = getInt32Ptr(primary_group.(int32))
		user.GroupCreate = getBoolPtr(false)
	}

	if groups, ok := d.GetOk("groups"); ok {
		user.Groups = expandIntegers(groups.(*schema.Set).List())
	}

	if sshpubkey, ok := d.GetOk("sshpubkey"); ok {
		user.Sshpubkey.Set(getStringPtr(sshpubkey.(string)))
	}

	return user, nil
}

func expandUserForUpdate(d *schema.ResourceData) (api.UpdateUserParams, error) {
	user := api.UpdateUserParams{
		// @TODO:
		HomeMode: getStringPtr("755"),
	}

	if full_name, ok := d.GetOk("full_name"); ok {
		user.FullName = getStringPtr(full_name.(string))
	}

	if username, ok := d.GetOk("username"); ok {
		user.Username = getStringPtr(username.(string))
	}

	if uid, ok := d.GetOk("uid"); ok {
		user.Uid = getInt32Ptr(int32(uid.(int)))
	}

	if home, ok := d.GetOk("home"); ok {
		user.Home = getStringPtr(home.(string))
	}

	if shell, ok := d.GetOk("shell"); ok {
		user.Shell = getStringPtr(shell.(string))
	}

	smb := d.Get("smb")
	if smb != nil {
		user.Smb = getBoolPtr(smb.(bool))
	}

	if d.HasChange("password") {
		password, password_set := d.GetOk("password")

		if password_set {
			user.Password = getStringPtr(password.(string))
		}

		user.PasswordDisabled = getBoolPtr(!password_set)
	}

	locked := d.Get("locked")
	if locked != nil {
		// Making a user locked needs a password set.
		if locked.(bool) && d.Get("password").(string) == "" {
			return user, errors.New("Locked needs the user to have a password.")
		}
		user.Locked = getBoolPtr(locked.(bool))
	}

	sudo := d.Get("sudo")
	if sudo != nil {
		user.Sudo = getBoolPtr(sudo.(bool))
	}

	sudo_nopasswd := d.Get("sudo_nopasswd")
	if sudo_nopasswd != nil {
		user.SudoNopasswd = getBoolPtr(sudo_nopasswd.(bool))
	}

	if sudo_commands, ok := d.GetOk("sudo_commands"); ok {
		user.SudoCommands = expandStrings(sudo_commands.(*schema.Set).List())
	}

	microsoft_account := d.Get("microsoft_account")
	if microsoft_account != nil {
		user.MicrosoftAccount = getBoolPtr(microsoft_account.(bool))
	}

	if email, ok := d.GetOk("email"); ok {
		user.Email.Set(getStringPtr(email.(string)))
	}

	if primary_group, ok := d.GetOk("primary_group"); ok {
		user.Group = getInt32Ptr(int32(primary_group.(int)))
	}

	if groups, ok := d.GetOk("groups"); ok {
		user.Groups = expandIntegers(groups.(*schema.Set).List())
	}

	if sshpubkey, ok := d.GetOk("sshpubkey"); ok {
		user.Sshpubkey.Set(getStringPtr(sshpubkey.(string)))
	}

	return user, nil
}

func expandUserForDelete(d *schema.ResourceData) api.DeleteUserParams {
	user := api.DeleteUserParams{
		DeleteGroup: getBoolPtr(false),
	}

	// Dumb heuristic: Delete the user's primary group if it has
	// the same name as the user, otherwise leave it empty.
	if primary_group_name, ok := d.GetOk("primary_group_name"); ok {
		user.DeleteGroup = getBoolPtr(primary_group_name.(string) == d.Get("username"))
	}

	return user
}
