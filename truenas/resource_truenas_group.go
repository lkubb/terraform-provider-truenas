package truenas

import (
	"context"
	api "github.com/dariusbakunas/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"regexp"
	"strconv"
)

func resourceTrueNASGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "Using groups in TrueNAS can be an efficient way of managing permissions for many similar user accounts.",
		CreateContext: resourceTrueNASGroupCreate,
		ReadContext:   resourceTrueNASGroupRead,
		UpdateContext: resourceTrueNASGroupUpdate,
		DeleteContext: resourceTrueNASGroupDelete,
		Schema: map[string]*schema.Schema{
			"group_id": &schema.Schema{
				Description: "Group ID",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"gid": &schema.Schema{
				Description: "GID",
				Type:        schema.TypeInt,
				Optional:    true,
				// This is arbitrary. Sensible? @TODO
				ValidateFunc: validation.IntBetween(0, 65535),
			},
			"name": &schema.Schema{
				Description: "This group's name.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 16),
					validation.StringDoesNotContainAny(`	,:+&#%^()!@~*?<>= `),
					validation.StringMatch(regexp.MustCompile(`^[^\$-][^\$\s]*.$`), "Username must not begin with a hyphen and `$` can only be included as the last character."),
				),
			},
			"builtin": &schema.Schema{
				Description: "Whether this is a builtin group.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"sudo": &schema.Schema{
				Description: "Permit sudo usage by this group.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"sudo_nopasswd": &schema.Schema{
				Description: "Permit sudo usage without password authentication for this group.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"sudo_commands": &schema.Schema{
				Description: "List of permitted sudo commands for this group. They must not rely on $PATH (i.e. have to be absolute).",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"smb": &schema.Schema{
				Description: "Samba authentication: Set to allow group to authenticate to Samba shares.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"users": &schema.Schema{
				Description: "List of User IDs (resouce IDs, not UID) in this group.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"local": &schema.Schema{
				Description: "Whether this is a local group.",
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

func resourceTrueNASGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id, err := strconv.Atoi(d.Id())

	resp, http, err := c.GroupApi.GetGroup(ctx, int32(id)).Execute()

	if err != nil {
		d.SetId("")
		if http.StatusCode == 404 {
			return nil
		}
		return diag.Errorf("error getting group: %s", err)
	}

	d.Set("name", resp.Group)

	if resp.Gid != nil {
		d.Set("gid", *resp.Gid)
	}

	d.Set("builtin", *resp.Builtin)
	d.Set("sudo", *resp.Sudo)
	d.Set("sudo_nopasswd", *resp.SudoNopasswd)

	if resp.SudoCommands != nil {
		if err := d.Set("sudo_commands", flattenStringList(*resp.SudoCommands)); err != nil {
			return diag.Errorf("error setting sudo_commands: %s", err)
		}
	}

	d.Set("smb", *resp.Smb)

	if resp.Users != nil {
		if err := d.Set("users", flattenInt32List(*resp.Users)); err != nil {
			return diag.Errorf("error setting users: %s", err)
		}
	}

	d.Set("local", *resp.Local)
	d.Set("id_type_both", *resp.IdTypeBoth)

	d.Set("group_id", int(resp.Id))

	return diags
}

func resourceTrueNASGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)

	group := expandGroup(d)

	resp, _, err := c.GroupApi.CreateGroup(ctx).CreateGroupParams(group).Execute()

	if err != nil {
		return diag.Errorf("error creating group: %s", err)
	}

	d.SetId(strconv.Itoa(int(resp)))

	return resourceTrueNASGroupRead(ctx, d, m)
}

func resourceTrueNASGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)

	group := expandGroupForDelete(d)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting TrueNAS group: %s", strconv.Itoa(id))

	_, err = c.GroupApi.DeleteGroup(ctx, int32(id)).DeleteGroupParams(group).Execute()

	if err != nil {
		return diag.Errorf("error deleting group: %s", err)
	}

	log.Printf("[INFO] TrueNAS group (%s) deleted", strconv.Itoa(id))
	d.SetId("")

	return diags
}

func resourceTrueNASGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*api.APIClient)
	group := expandGroup(d)

	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	_, _, err = c.GroupApi.UpdateGroup(ctx, int32(id)).CreateGroupParams(group).Execute()

	if err != nil {
		return diag.Errorf("error updating group: %s", err)
	}

	return resourceTrueNASGroupRead(ctx, d, m)
}

func expandGroup(d *schema.ResourceData) api.CreateGroupParams {
	group := api.CreateGroupParams{
		Name:              d.Get("name").(string),
		AllowDuplicateGid: getBoolPtr(false), // @TODO?
	}

	if gid, ok := d.GetOk("gid"); ok {
		group.Gid = getInt32Ptr(int32(gid.(int)))
	}

	smb := d.Get("smb")
	if smb != nil {
		group.Smb = getBoolPtr(smb.(bool))
	}

	sudo := d.Get("sudo")
	if sudo != nil {
		group.Sudo = getBoolPtr(sudo.(bool))
	}

	sudo_nopasswd := d.Get("sudo_nopasswd")
	if sudo_nopasswd != nil {
		group.SudoNopasswd = getBoolPtr(sudo_nopasswd.(bool))
	}

	if sudo_commands, ok := d.GetOk("sudo_commands"); ok {
		group.SudoCommands = expandStrings(sudo_commands.(*schema.Set).List())
	}

	if users, ok := d.GetOk("users"); ok {
		group.Users = expandIntegers(users.(*schema.Set).List())
	}

	return group
}

func expandGroupForDelete(d *schema.ResourceData) api.DeleteGroupParams {
	group := api.DeleteGroupParams{
		DeleteUsers: getBoolPtr(false),
	}

	return group
}
