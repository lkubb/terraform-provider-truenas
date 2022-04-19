resource "truenas_group" "group" {
  name = "fsociety"
  gid = 1337
  sudo = true
  sudo_nopasswd = true
  sudo_commands = [
    "/usr/bin/true"
  ]
  smb = true
  users = [
    1,
  ]
}
