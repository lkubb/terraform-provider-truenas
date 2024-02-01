package truenas

import (
	"context"
	api "github.com/dariusbakunas/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
)

func dataSourceTrueNASUser() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a specific user",
		ReadContext: dataSourceTrueNASUserRead,
		Schema: map[string]*schema.Schema{
			"user_id": &schema.Schema{
				Description: "User ID",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"uid": &schema.Schema{
				Description: "UID",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"username": &schema.Schema{
				Description: "Username",
				Type:        schema.TypeString,
				Computed:    true,
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
				Description: "This user's home directory.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"shell": &schema.Schema{
				Description: "This user's shell executable.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"full_name": &schema.Schema{
				Description: "This user's full name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"builtin": &schema.Schema{
				Description: "Whether this is a builtin user.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"smb": &schema.Schema{
				Description: "Samba authentication",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"password_disabled": &schema.Schema{
				Description: "Password disabled: The account cannot use password-based logins for services.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"locked": &schema.Schema{
				Description: "Prevents the user from logging in or using password-based services until this option is unset.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"sudo": &schema.Schema{
				Description: "Permits sudo usage by this user.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"sudo_nopasswd": &schema.Schema{
				Description: "Permits sudo usage without password authentication for this user.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"sudo_commands": &schema.Schema{
				Description: "List of permitted sudo commands for this user.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"microsoft_account": &schema.Schema{
				Description: "Microsoft account: Allows additional username authentication methods for Windows >=8 clients.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			// "attributes": &schema.Schema{
			// 	Description: "Attributes",
			// 	Type:        schema.TypeMap,
			// 	Computed:    true,
			// },
			"email": &schema.Schema{
				Description: "This user's email address.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"primary_group": &schema.Schema{
				Description: "Group ID of this user's primary group (resource ID, not GID).",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"primary_group_name": &schema.Schema{
				Description: "The name of this user's primary group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"groups": &schema.Schema{
				Description: "List of group IDs this user is member in (resource IDs, not GIDs).",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"sshpubkey": &schema.Schema{
				Description: "This user's SSH pubkey.",
				Type:        schema.TypeString,
				Computed:    true,
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

func dataSourceTrueNASUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id := d.Get("user_id").(int)

	resp, _, err := c.UserApi.GetUser(ctx, int32(id)).Execute()

	if err != nil {
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

	if resp.Email.IsSet() {
		d.Set("email", *resp.Email.Get())
	}

	if resp.Group != nil {
		if resp.Group.Id != nil {
			d.Set("primary_group", *resp.Group.Id)
		}

		if resp.Group.BsdgrpGroup != nil {
			d.Set("primary_group_name", *resp.Group.BsdgrpGroup)
		}
	}

	if err := d.Set("groups", flattenInt32List(resp.Groups)); err != nil {
		return diag.Errorf("error setting users: %s", err)
	}

	if resp.Sshpubkey.IsSet() {
		d.Set("sshpubkey", *resp.Sshpubkey.Get())
	}

	d.Set("local", *resp.Local)
	d.Set("id_type_both", *resp.IdTypeBoth)

	d.SetId(strconv.Itoa(int(resp.Id)))

	return diags
}
