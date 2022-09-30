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
