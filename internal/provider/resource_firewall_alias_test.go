package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestFirewallAliasResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccFirewallAliasResourceConfig("pmve_firewall_alias_test", "1.2.3.0/24", "this is a comment"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_firewall_alias.test", "name", "pmve_firewall_alias_test"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_alias.test", "cidr", "1.2.3.0/24"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_alias.test", "comment", "this is a comment"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "proxmoxve_firewall_alias.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccFirewallAliasResourceConfig("pmve_firewall_alias_test", "4.5.0.0/16", "no comments"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_firewall_alias.test", "id", "pmve_firewall_alias_test"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_alias.test", "name", "pmve_firewall_alias_test"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_alias.test", "cidr", "4.5.0.0/16"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_alias.test", "comment", "no comments"),
				),
			},
			// Rename and read
			{
				Config: testAccFirewallAliasResourceConfig("pmve_firewall_alias_test_renamed", "4.5.0.0/16", "no comments"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_firewall_alias.test", "id", "pmve_firewall_alias_test_renamed"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_alias.test", "name", "pmve_firewall_alias_test_renamed"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_alias.test", "cidr", "4.5.0.0/16"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_alias.test", "comment", "no comments"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccFirewallAliasResourceConfig(name, cidr, comment string) string {
	return fmt.Sprintf(`
		resource "proxmoxve_firewall_alias" "test" {
			name    = "%s"
			cidr    = "%s"
			comment = "%s"
		}
	`, name, cidr, comment)
}
