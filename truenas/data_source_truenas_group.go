package truenas

import (
	"context"
	api "github.com/dariusbakunas/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
)

func dataSourceTrueNASGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a specific group",
		ReadContext: dataSourceTrueNASGroupRead,
		Schema: map[string]*schema.Schema{
			"group_id": &schema.Schema{
				Description: "Group ID",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"gid": &schema.Schema{
				Description: "GID",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"name": &schema.Schema{
				Description: "This group's name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"builtin": &schema.Schema{
				Description: "Whether this is a builtin group.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"sudo": &schema.Schema{
				Description: "Permits sudo usage by this group.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"sudo_nopasswd": &schema.Schema{
				Description: "Permit sudo usage without password authentication for this group.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"sudo_commands": &schema.Schema{
				Description: "List of permitted sudo commands for this group.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"smb": &schema.Schema{
				Description: "Samba authentication",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"users": &schema.Schema{
				Description: "List of Group IDs (resouce IDs, not GID) in this group.",
				Type:        schema.TypeSet,
				Computed:    true,
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

func dataSourceTrueNASGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*api.APIClient)
	id := d.Get("group_id").(int)

	resp, _, err := c.GroupApi.GetGroup(ctx, int32(id)).Execute()

	if err != nil {
		return diag.Errorf("error getting group: %s", err)
	}

	// for whatever reason, the group name is called `group`
	d.Set("name", resp.Group)

	if resp.Gid != nil {
		d.Set("gid", *resp.Gid)
	}

	d.Set("builtin", *resp.Builtin)
	d.Set("sudo", *resp.Sudo)
	d.Set("sudo_nopasswd", *resp.SudoNopasswd)

	if err := d.Set("sudo_commands", flattenStringList(resp.SudoCommands)); err != nil {
		return diag.Errorf("error setting sudo_commands: %s", err)
	}

	d.Set("smb", *resp.Smb)

	if err := d.Set("users", flattenInt32List(resp.Users)); err != nil {
		return diag.Errorf("error setting users: %s", err)
	}

	d.Set("local", *resp.Local)
	d.Set("id_type_both", *resp.IdTypeBoth)

	d.SetId(strconv.Itoa(int(resp.Id)))

	return diags
}
