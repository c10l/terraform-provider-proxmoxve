terraform {
  required_providers {
    proxmoxve = {
      source = "c10l/proxmoxve"
    }
  }
}

provider "proxmoxve" {
  base_url      = "https://pmve.example.com:8006"
  token_id      = "apiuser@pam!proxmoxve_terraform_token"
  secret        = "token_secret"
  root_password = "password"
  tls_insecure  = false
}
