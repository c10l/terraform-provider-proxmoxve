# This resource is for declaring an empty Security Group, to be populated by the ancillary
# proxmoxve_firewall_group_rule resource.
resource "proxmoxve_firewall_group" "management" {
  name    = "sec_group"
  comment = "this is an optional comment"
}
