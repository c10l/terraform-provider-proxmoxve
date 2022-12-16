# List all firewall_refs
data "proxmoxve_firewall_refs" "all_refs" {}

# List only firewall_refs of type 'alias'
data "proxmoxve_firewall_refs" "alias" {
  type = "alias"
}

# List only firewall_refs of type 'ipset'
data "proxmoxve_firewall_refs" "ipset" {
  type = "ipset"
}
