package truenas

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceTruenasUser_basic(t *testing.T) {
	resourceName := "data.truenas_user.user"

	suffix := acctest.RandStringFromCharSet(3, acctest.CharSetAlphaNum)
	userName := fmt.Sprintf("%s_%s", testResourcePrefix, suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResourceTruenasUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataSourceTruenasUserConfig(userName, testPoolName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", userName),
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

func testAccCheckDataSourceTruenasUserConfig(userName string, testPoolName string) string {
	return fmt.Sprintf(`
	resource "truenas_user" "usertest" {
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

	data "truenas_user" "user" {
		user_id = resource.truenas_user.usertest.user_id
	}
	`, userName, testPoolName, userName)
}
