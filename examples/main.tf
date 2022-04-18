terraform {
  required_providers {
    proxmoxve = {
      version = "0.0.1"
      source  = "c10l.cc/local/proxmoxve"
    }
  }
}

provider "proxmoxve" {}

data "proxmoxve_version" "current" {}

output "version" {
  value = data.proxmoxve_version.current
}
