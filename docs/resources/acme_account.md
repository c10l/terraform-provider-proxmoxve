---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "proxmoxve_acme_account Resource - terraform-provider-proxmoxve"
subcategory: ""
description: |-
  Define an ACME account with CA.NOTE: This resource requires the provider attribute root_password or the environment variable PROXMOXVE_ROOT_PASSWORD set.
---

# proxmoxve_acme_account (Resource)

Define an ACME account with CA.<p />**NOTE:** This resource requires the provider attribute `root_password` or the environment variable `PROXMOXVE_ROOT_PASSWORD` set.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `contact` (String) Contact email addresses.
- `name` (String) Name of the account in Proxmox VE.

### Optional

- `directory` (String) URL of ACME CA directory endpoint. Defaults to the production directory of Let's Encrypt.
- `tos_url` (String) URL of CA TermsOfService - setting this indicates agreement.

### Read-Only

- `id` (String) The ID of this resource.


