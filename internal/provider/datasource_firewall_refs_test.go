package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceFirewallRefs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccDataSourceFirewallRefsConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("data.proxmoxve_firewall_refs.test", "type"),
					resource.TestCheckTypeSetElemNestedAttrs("data.proxmoxve_firewall_refs.test", "refs.*", map[string]string{"name": "proxmoxve_firewall_refs_test"}),
				),
			},
		},
	})
}

const testAccDataSourceFirewallRefsConfig = `
		resource "proxmoxve_firewall_ipset" "test" {
			name = "proxmoxve_firewall_refs_test"
		}

		data "proxmoxve_firewall_refs" "test" {
			depends_on = [proxmoxve_firewall_ipset.test]
		}
`
