package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceFirewallAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccDataSourceFirewallAliasConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.proxmoxve_firewall_alias.test", "name", "proxmoxve_firewall_alias_datasource_test"),
					resource.TestCheckResourceAttr("data.proxmoxve_firewall_alias.test", "cidr", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("data.proxmoxve_firewall_alias.test", "comment", "foobar"),
				),
			},
		},
	})
}

const testAccDataSourceFirewallAliasConfig = `resource "proxmoxve_firewall_alias" "test" {
		name    = "proxmoxve_firewall_alias_datasource_test"
		cidr    = "0.0.0.0/0"
		comment = "foobar"
	}

	data "proxmoxve_firewall_alias" "test" {
		name = resource.proxmoxve_firewall_alias.test.name
	}`
