---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "truenas_share_smb Data Source - terraform-provider-truenas"
subcategory: ""
description: |-
  Get information about specific SMB share
---

# truenas_share_smb (Data Source)

Get information about specific SMB share

## Example Usage

```terraform
data "truenas_share_smb" "smb" {
  sharesmb_id = 1
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `sharesmb_id` (Number) SMB Share ID

### Optional

- `id` (String) The ID of this resource.

### Read-Only

- `aapl_name_mangling` (Boolean) Use Apple-style Character Encoding
- `abe` (Boolean) Access based share enumeration
- `acl` (Boolean) Enable support for storing the SMB Security Descriptor as a Filesystem ACL
- `auxsmbconf` (String) Auxiliary smb4.conf parameters
- `browsable` (Boolean) Browsable to network clients
- `comment` (String) Any notes about this SMB share
- `durablehandle` (Boolean) Enable SMB2/3 Durable Handles: Allow using open file handles that can withstand short disconnections
- `enabled` (Boolean) Enable this share
- `fsrvp` (Boolean) Enable support for the File Server Remote VSS Protocol.
- `guestok` (Boolean) Allow access to this share without a password
- `home` (Boolean) Use as home share
- `hostsallow` (Set of String) Authorized hosts (IP/hostname)
- `hostsdeny` (Set of String) Disallowed hosts (IP/hostname). Pass 'ALL' to use whitelist model.
- `locked` (Boolean) Locking status of this share
- `name` (String) SMB share name
- `path` (String) Path to shared directory
- `path_suffix` (String) Append a suffix to the share connection path. This is used to provide unique shares on a per-user, per-computer, or per-IP address basis.
- `purpose` (String) You can set a share purpose to apply and lock pre-determined advanced options for the share.
- `recyclebin` (Boolean) Export recycle bin
- `ro` (Boolean) Prohibit writing
- `shadowcopy` (Boolean) Export ZFS snapshots as Shadow Copies for Microsoft Volume Shadow Copy Service (VSS) clients
- `streams` (Boolean) Enable Alternate Data Streams: Allow multiple NTFS data streams. Disabling this option causes macOS to write streams to files on the filesystem.
- `timemachine` (Boolean) Enable TimeMachine backups to this share
- `vuid` (String) Share VUID (set when using as TimeMachine share)

