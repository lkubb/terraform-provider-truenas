resource "truenas_user" "user" {
  uid = 1337
  username = "elliot"
  home = "/mnt/tank/home/elliot"
  shell = "/usr/local/bin/zsh"
  full_name = "Mr Robot"
  smb = true
  password = "5uperS3(ur3!"
  locked = false
  sudo = true
  sudo_nopasswd = true
  sudo_commands = [
    "/usr/bin/whoami"
  ]
  microsoft_account = false
  email = "elliot.alderson@protonmail.ch"
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
