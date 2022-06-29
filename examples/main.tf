terraform {
  required_providers {
    proxmoxve = {
      version = "0.0.1"
      source  = "github.com/c10l/proxmoxve"
    }
  }
}

provider "proxmoxve" {
  base_url     = "https://10.100.10.251:8006"
  token_id     = "root@pam!tests"
  secret       = "c9d6c337-d14f-469f-8487-1a69b9b3118d"
  tls_insecure = true
}

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
