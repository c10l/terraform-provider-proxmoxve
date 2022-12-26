---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "proxmoxve_firewall_ipset_cidr Resource - terraform-provider-proxmoxve"
subcategory: ""
description: |-
  
---

# proxmoxve_firewall_ipset_cidr (Resource)



## Example Usage

```terraform
resource "proxmoxve_firewall_ipset_cidr" "example" {
  ipset_name = "ipset_name"
  cidr       = "10.0.1.0/24"
  comment    = "some_comment"
}

resource "proxmoxve_firewall_ipset_cidr" "negated" {
  ipset_name = "ipset_name"
  cidr       = "2a01::/16"
  no_match   = true
  comment    = "block the Internet!"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cidr` (String) CIDR to be configured. e.g. `10.0.0.0/8`, `fd65::/16`.
- `ipset_name` (String) Name of the IPSet on which to attach this CIDR.

### Optional

- `comment` (String)
- `no_match` (Boolean) Set to `true` to negate the CIDR rather than matching it.

### Read-Only

- `id` (String) The ID of this resource.

