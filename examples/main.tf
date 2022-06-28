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

data "proxmoxve_storage" "local" {
  storage = "local"
}

output "version" {
  value = data.proxmoxve_version.current
}

output "storage" {
  value = data.proxmoxve_storage.local
}
