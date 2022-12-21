# Get firewall_alias named `lan`
data "proxmoxve_firewall_alias" "alias" {
  name = "lan"
}
