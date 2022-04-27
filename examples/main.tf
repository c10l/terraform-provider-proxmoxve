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

resource "proxmoxve_pool" "checkmeout" {
  pool_id = "checkychecky"
  comment = data.proxmoxve_version.current.version
}

resource "proxmoxve_pool" "fromtheother" {
  pool_id = "theother"
  comment = "updateme ${proxmoxve_pool.checkmeout.pool_id}"
}

resource "proxmoxve_pool" "foofoofaafaa" {
  pool_id = "foofoofaafaa"
  comment = "something else really"
}

output "version" {
  value = data.proxmoxve_version.current
}
