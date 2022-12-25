package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestFirewallIPSetResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccFirewallIPSetResourceConfig("pmve_firewall_ipset_test", "this is a comment"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_firewall_ipset.test", "name", "pmve_firewall_ipset_test"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_ipset.test", "comment", "this is a comment"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "proxmoxve_firewall_ipset.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccFirewallIPSetResourceConfig("pmve_firewall_ipset_test", "no comments"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_firewall_ipset.test", "id", "pmve_firewall_ipset_test"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_ipset.test", "name", "pmve_firewall_ipset_test"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_ipset.test", "comment", "no comments"),
				),
			},
			// Rename and read
			{
				Config: testAccFirewallIPSetResourceConfig("pmve_firewall_ipset_test_renamed", "no comments"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_firewall_ipset.test", "id", "pmve_firewall_ipset_test_renamed"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_ipset.test", "name", "pmve_firewall_ipset_test_renamed"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_ipset.test", "comment", "no comments"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccFirewallIPSetResourceConfig(name, comment string) string {
	return fmt.Sprintf(`
		resource "proxmoxve_firewall_ipset" "test" {
			name    = "%s"
			comment = "%s"
		}
	`, name, comment)
}
