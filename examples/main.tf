terraform {
  required_providers {
    proxmoxve = {
      version = "0.0.1"
      source  = "github.com/c10l/proxmoxve"
    }
  }
}

provider "proxmoxve" {}

data "proxmoxve_version" "current" {}

data "proxmoxve_storage" "local" {
  storage = "local"
}

resource "proxmoxve_storage_dir" "test" {
  storage = "test"
  path    = "/test"
}

output "version" {
  value = data.proxmoxve_version.current
}

output "storage" {
  value = data.proxmoxve_storage.local
}
