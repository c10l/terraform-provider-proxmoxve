terraform {
  required_providers {
    proxmoxve = {
      version = "0.0.1"
      source  = "github.com/c10l/proxmoxve"
    }
  }
}

resource "proxmoxve_acme_account" "test" {
  name      = "prettyaccount"
  contact   = "foobar@baz.com"
  directory = "https://127.0.0.1:14000/dir"
  tos_url   = "foobar"
}

resource "proxmoxve_firewall_ipset" "management" {
  name    = "management"
  comment = "this is an optional comment"
}

resource "proxmoxve_firewall_ipset_cidr" "infra_workstations" {
  ipset_name = proxmoxve_firewall_ipset.management.name
  cidr       = "192.168.10.0/24"
  comment    = "this is the CIDR of the admin workstations network"
}
