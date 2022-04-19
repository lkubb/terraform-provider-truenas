package truenas

import (
	"context"
	"fmt"
	api "github.com/dariusbakunas/truenas-go-sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestAccResourceTruenasGroup_basic(t *testing.T) {
	resourceName := "truenas_group.group"

	var group api.Group

	suffix := acctest.RandStringFromCharSet(3, acctest.CharSetAlphaNum)
	groupName := fmt.Sprintf("%s_%s", testResourcePrefix, suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResourceTruenasGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourceTruenasGroupConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					// make sure the API reports the expected values
					testAccCheckTruenasGroupResourceExists(resourceName, &group),
					testAccCheckTruenasGroupResourceAttributes(t, resourceName, &group, groupName),
					// make sure the Terraform resource matches
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

func testAccCheckResourceTruenasGroupConfig(groupName string) string {
	return fmt.Sprintf(`
	resource "truenas_group" "group" {
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
	`, groupName)
}

func testAccCheckTruenasGroupResourceExists(n string, group *api.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("group resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no group resource ID is set")
		}

		client := testAccProvider.Meta().(*api.APIClient)

		id, err := strconv.Atoi(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not convert resource ID of group: %s", n)
		}

		resp, _, err := client.GroupApi.GetGroup(context.Background(), int32(id)).Execute()

		if err != nil {
			return err
		}

		if strconv.Itoa(int(resp.Id)) != rs.Primary.ID {
			return fmt.Errorf("group not found")
		}

		*group = resp
		return nil
	}
}

func testAccCheckTruenasGroupResourceAttributes(t *testing.T, n string, group *api.Group, groupName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if group.Group != groupName {
			return fmt.Errorf("remote name for group does not match expected")
		}

		if *group.Gid != 17357 {
			return fmt.Errorf("remote GID for group does not match expected")
		}

		if *group.Sudo != true {
			return fmt.Errorf("remote sudo for group does not match expected")
		}

		if *group.SudoNopasswd != true {
			return fmt.Errorf("remote sudo_nopasswd for group does not match expected")
		}

		// order does not matter
		if !assert.ElementsMatch(t, *group.SudoCommands, []string{"/usr/bin/id"}) {
			return fmt.Errorf("remote sudo_commands for group do not match expected")
		}

		if *group.Smb != false {
			return fmt.Errorf("remote smb for group does not match expected")
		}

		if !assert.ElementsMatch(t, *group.Users, []int32{1}) {
			return fmt.Errorf("remote users for group do not match expected")
		}

		return nil
	}
}

func testAccCheckResourceTruenasGroupDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*api.APIClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "truenas_group" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not convert ID of group: %s", rs.Primary.ID)
		}

		_, r, err := client.GroupApi.GetGroup(context.Background(), int32(id)).Execute()

		if err == nil {
			return fmt.Errorf("group (%s) still exists", rs.Primary.ID)
		}

		// check if error is in fact 404 (not found)
		if r.StatusCode != 404 {
			return fmt.Errorf("Error occured while checking for absence of group (%s)", rs.Primary.ID)
		}
	}

	return nil
}
