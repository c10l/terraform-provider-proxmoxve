package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestFirewallGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccFirewallGroupResourceConfig("pmve_fw_group_test", "this is a comment"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_firewall_group.test", "name", "pmve_fw_group_test"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_group.test", "comment", "this is a comment"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "proxmoxve_firewall_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccFirewallGroupResourceConfig("pmve_fw_group_test", "no comments"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_firewall_group.test", "id", "pmve_fw_group_test"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_group.test", "name", "pmve_fw_group_test"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_group.test", "comment", "no comments"),
				),
			},
			// Rename and read
			{
				Config: testAccFirewallGroupResourceConfig("pmve_fw_group_ren", "no comments"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_firewall_group.test", "id", "pmve_fw_group_ren"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_group.test", "name", "pmve_fw_group_ren"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_group.test", "comment", "no comments"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccFirewallGroupResourceConfig(name, comment string) string {
	return fmt.Sprintf(`
		resource "proxmoxve_firewall_group" "test" {
			name    = "%s"
			comment = "%s"
		}

		resource "proxmoxve_firewall_group" "test_no_comment" {
			name = "pmve_fw_group_nocm"
		}
	`, name, comment)
}
