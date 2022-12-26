resource "proxmoxve_firewall_ipset_cidr" "example" {
  ipset_name = "ipset_name"
  cidr       = "10.0.1.0/24"
  comment    = "some_comment"
}

resource "proxmoxve_firewall_ipset_cidr" "negated" {
  ipset_name = "ipset_name"
  cidr       = "2a01::/16"
  no_match   = true
  comment    = "block the Internet!"
}
