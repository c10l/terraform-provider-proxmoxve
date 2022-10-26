resource "proxmoxve_acme_account" "test" {
  name      = "acme_provider"
  contact   = "foo@example.com"
  directory = "https://acme_provider.example.com/dir"
  tos_url   = "https://acme_provider.example.com/terms_of_service.pdf"
}

resource "proxmoxve_acme_account" "letsencrypt_staging" {
  name      = "letsencrypt_staging"
  contact   = "foo@example.com"
  directory = "https://acme-staging-v02.api.letsencrypt.org/directory"
  tos_url   = "https://letsencrypt.org/documents/LE-SA-v1.2-November-15-2017.pdf"
}

resource "proxmoxve_acme_account" "letsencrypt_prod" {
  name      = "letsencrypt_prod"
  contact   = "foo@example.com"
  directory = "https://acme-v02.api.letsencrypt.org/directory"
  tos_url   = "https://letsencrypt.org/documents/LE-SA-v1.2-November-15-2017.pdf"
}
