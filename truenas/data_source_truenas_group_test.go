package truenas

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceTruenasGroup_basic(t *testing.T) {
	resourceName := "data.truenas_group.group"

	suffix := acctest.RandStringFromCharSet(3, acctest.CharSetAlphaNum)
	groupName := fmt.Sprintf("%s_%s", testResourcePrefix, suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResourceTruenasGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataSourceTruenasGroupConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "gid", "17357"),
					resource.TestCheckResourceAttr(resourceName, "sudo", "true"),
					resource.TestCheckResourceAttr(resourceName, "sudo_nopasswd", "true"),
					resource.TestCheckTypeSetElemAttr(resourceName, "sudo_commands.*", "/usr/bin/id"),
					resource.TestCheckResourceAttr(resourceName, "smb", "false"),
					resource.TestCheckTypeSetElemAttr(resourceName, "users.*", "1"),
				),
			},
		},
	})
}

func testAccCheckDataSourceTruenasGroupConfig(groupName string) string {
	return fmt.Sprintf(`
	resource "truenas_group" "grouptest" {
		name = "%s"
		gid = 17357
		sudo = true
		sudo_nopasswd = true
		sudo_commands = [
			"/usr/bin/id"
		]
		smb = false
		users = [
			1,
		]
	}

	data "truenas_group" "group" {
		group_id = resource.truenas_group.grouptest.group_id
	}
	`, groupName)
}
