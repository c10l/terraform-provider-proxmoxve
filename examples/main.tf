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
  storage = "dir_test"
  path    = "/test"
}

resource "proxmoxve_storage_btrfs" "test" {
  storage = "btrfs_test"
  path    = "/foo/bar"
  disable = true
}

resource "proxmoxve_storage_nfs" "test" {
  storage = "nfs_test"
  server  = "1.2.3.4"
  export  = "/test"
  disable = true
}

output "version" {
  value = data.proxmoxve_version.current
}

output "storage" {
  value = data.proxmoxve_storage.local
}
