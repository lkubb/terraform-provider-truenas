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

func TestAccResourceTruenasUser_basic(t *testing.T) {
	resourceName := "truenas_user.user"

	var user api.User

	suffix := acctest.RandStringFromCharSet(3, acctest.CharSetAlphaNum)
	userName := fmt.Sprintf("%s_%s", testResourcePrefix, suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResourceTruenasUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourceTruenasUserConfigSimple(userName, testPoolName),
				Check: resource.ComposeTestCheckFunc(
					// make sure the API reports the expected values
					testAccCheckTruenasUserResourceExists(resourceName, &user),
					testAccCheckTruenasUserResourceAttributes(t, resourceName, &user, userName, testPoolName),
					// make sure the Terraform resource matches
					resource.TestCheckResourceAttr(resourceName, "username", userName),
					resource.TestCheckResourceAttr(resourceName, "primary_group_name", userName),
					resource.TestCheckResourceAttr(resourceName, "uid", "17357"),
					resource.TestCheckResourceAttr(resourceName, "home", fmt.Sprintf("/mnt/%s/%s", testPoolName, userName)),
					resource.TestCheckResourceAttr(resourceName, "shell", "/usr/sbin/nologin"),
					resource.TestCheckResourceAttr(resourceName, "full_name", "Terraform Test"),
					resource.TestCheckResourceAttr(resourceName, "smb", "false"),
					resource.TestCheckResourceAttr(resourceName, "locked", "true"),
					resource.TestCheckResourceAttr(resourceName, "sudo", "true"),
					resource.TestCheckResourceAttr(resourceName, "sudo_nopasswd", "true"),
					resource.TestCheckTypeSetElemAttr(resourceName, "sudo_commands.*", "/usr/bin/id"),
					resource.TestCheckResourceAttr(resourceName, "microsoft_account", "true"),
					resource.TestCheckResourceAttr(resourceName, "email", "tf@test.acc"),
					resource.TestCheckTypeSetElemAttr(resourceName, "groups.*", "11"),
					resource.TestCheckResourceAttr(resourceName, "sshpubkey", `ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAklOUpkDHrfHY17SbrmTIpNLTGK9Tjom/BWDSU
GPl+nafzlHDTYW7hdI4yZ5ew18JH4JW9jbhUFrviQzM7xlELEVf4h9lFX5QVkbPppSwg0cda3
Pbv7kOdJ/MTyBlWXFCR+HAo3FXRitBqxiX1nKhXpHAZsMciLq8V6RjsNAQwdsdMFvSlVK/7XA
t3FaoJoAsncM1Q9x5+3V0Ww68/eIFmb1zuUFljQJKprrX88XypNDvjYNby6vw/Pb0rwert/En
mZ+AW4OZPnTPI89ZPmVMLuayrD2cE86Z/il8b+gw3r3+1nKatmIkjn2so1d01QraTlMqVSsbx
NrRFi9wrf+M7Q==
`),
				),
			},
		},
	})
}

func testAccCheckResourceTruenasUserConfigSimple(userName string, testPoolName string) string {
	return fmt.Sprintf(`
	resource "truenas_user" "user" {
		uid = 17357
		username = "%s"
		home = "/mnt/%s/%s"
		shell = "/usr/sbin/nologin"
		full_name = "Terraform Test"
		smb = false
		password = "5uperS3(ur3!"
		locked = true
		sudo = true
		sudo_nopasswd = true
		sudo_commands = [
			"/usr/bin/id"
		]
		microsoft_account = true
		email = "tf@test.acc"
		groups = [
			11,
		]
		sshpubkey = <<-EOF
			ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAklOUpkDHrfHY17SbrmTIpNLTGK9Tjom/BWDSU
			GPl+nafzlHDTYW7hdI4yZ5ew18JH4JW9jbhUFrviQzM7xlELEVf4h9lFX5QVkbPppSwg0cda3
			Pbv7kOdJ/MTyBlWXFCR+HAo3FXRitBqxiX1nKhXpHAZsMciLq8V6RjsNAQwdsdMFvSlVK/7XA
			t3FaoJoAsncM1Q9x5+3V0Ww68/eIFmb1zuUFljQJKprrX88XypNDvjYNby6vw/Pb0rwert/En
			mZ+AW4OZPnTPI89ZPmVMLuayrD2cE86Z/il8b+gw3r3+1nKatmIkjn2so1d01QraTlMqVSsbx
			NrRFi9wrf+M7Q==
			EOF
	}
	`, userName, testPoolName, userName)
}

func testAccCheckTruenasUserResourceExists(n string, user *api.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("user resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no user resource ID is set")
		}

		client := testAccProvider.Meta().(*api.APIClient)

		id, err := strconv.Atoi(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not convert ID of user: %s", n)
		}

		resp, _, err := client.UserApi.GetUser(context.Background(), int32(id)).Execute()

		if err != nil {
			return err
		}

		if strconv.Itoa(int(resp.Id)) != rs.Primary.ID {
			return fmt.Errorf("user not found")
		}

		*user = resp
		return nil
	}
}

func testAccCheckTruenasUserResourceAttributes(t *testing.T, n string, user *api.User, userName string, testPoolName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if user.Username != userName {
			return fmt.Errorf("remote username for user does not match expected")
		}

		if *user.Group.BsdgrpGroup != userName {
			return fmt.Errorf("remote primary_group_name for user does not match expected")
		}

		if *user.Uid != 17357 {
			return fmt.Errorf("remote UID for user does not match expected")
		}

		if *user.Home != fmt.Sprintf("/mnt/%s/%s", testPoolName, userName) {
			return fmt.Errorf("remote home for user does not match expected")
		}

		if *user.Shell != "/usr/sbin/nologin" {
			return fmt.Errorf("remote shell for user does not match expected")
		}

		if user.FullName != "Terraform Test" {
			return fmt.Errorf("remote full_name for user does not match expected")
		}

		if *user.Smb != false {
			return fmt.Errorf("remote smb for user does not match expected")
		}

		if *user.Locked != true {
			return fmt.Errorf("remote locked for user does not match expected")
		}

		if *user.Sudo != true {
			return fmt.Errorf("remote sudo for user does not match expected")
		}

		if *user.SudoNopasswd != true {
			return fmt.Errorf("remote sudo_nopasswd for user does not match expected")
		}

		// order does not matter
		if !assert.ElementsMatch(t, *user.SudoCommands, []string{"/usr/bin/id"}) {
			return fmt.Errorf("remote sudo_commands for user do not match expected")
		}

		if *user.MicrosoftAccount != true {
			return fmt.Errorf("remote microsoft_account for user does not match expected")
		}

		if *user.Email.Get() != "tf@test.acc" {
			return fmt.Errorf("remote email for user does not match expected")
		}

		if !assert.ElementsMatch(t, *user.Groups, []int32{11}) {
			return fmt.Errorf("remote groups for user do not match expected")
		}

		if *user.Sshpubkey.Get() != `ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAklOUpkDHrfHY17SbrmTIpNLTGK9Tjom/BWDSU
GPl+nafzlHDTYW7hdI4yZ5ew18JH4JW9jbhUFrviQzM7xlELEVf4h9lFX5QVkbPppSwg0cda3
Pbv7kOdJ/MTyBlWXFCR+HAo3FXRitBqxiX1nKhXpHAZsMciLq8V6RjsNAQwdsdMFvSlVK/7XA
t3FaoJoAsncM1Q9x5+3V0Ww68/eIFmb1zuUFljQJKprrX88XypNDvjYNby6vw/Pb0rwert/En
mZ+AW4OZPnTPI89ZPmVMLuayrD2cE86Z/il8b+gw3r3+1nKatmIkjn2so1d01QraTlMqVSsbx
NrRFi9wrf+M7Q==
` {
			return fmt.Errorf("remote sshpubkey for user does not match expected: \n'%s'", *user.Sshpubkey.Get())
		}

		return nil
	}
}

func testAccCheckResourceTruenasUserDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*api.APIClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "truenas_user" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("could not convert ID of user: %s", rs.Primary.ID)
		}

		_, r, err := client.UserApi.GetUser(context.Background(), int32(id)).Execute()

		if err == nil {
			return fmt.Errorf("user (%s) still exists", rs.Primary.ID)
		}

		if r.StatusCode != 404 {
			return fmt.Errorf("Error occured while checking for absence of user (%s)", rs.Primary.ID)
		}
	}

	return nil
}
