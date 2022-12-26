# This resource is for declaring an empty IPSet, to be populated by the ancillary
# proxmoxve_firewall_ipset_cidr resource.
resource "proxmoxve_firewall_ipset" "management" {
  name    = "management"
  comment = "this is an optional comment"
}

resource "proxmoxve_firewall_ipset_cidr" "infra_workstations" {
  ipset_name = proxmoxve_firewall_ipset.management.name
  cidr       = "192.168.10.0/24"
  comment    = "this is the CIDR of the admin workstations network"
}

resource "proxmoxve_firewall_ipset_cidr" "infra_cicd" {
  ipset_name = proxmoxve_firewall_ipset.management.name
  cidr       = "10.100.0.0/24"
  comment    = "this is the CIDR of the CI/CD server that runs Terraform and infra-as-code to manage the PVE cluster"
}
