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
  id      = "checkychecky"
  comment = data.proxmoxve_version.current.version
}

resource "proxmoxve_pool" "fromtheother" {
  id      = "theother"
  comment = "updateme ${proxmoxve_pool.checkmeout.id}"
}

resource proxmoxve_pool foofoofaafaa {
  id      = "foofoofaafaa"
  comment = "something else really"
}

output "version" {
  value = data.proxmoxve_version.current
}
