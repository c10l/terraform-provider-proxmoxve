package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
)

func TestSplitID(t *testing.T) {
	r := new(FirewallIPSetCIDRResource)
	ipSetName, cidr := r.splitID("foo/10.0.0.0/24")
	assert.Equal(t, "foo", ipSetName)
	assert.Equal(t, "10.0.0.0/24", cidr)
}

func TestFirewallIPSetCIDRResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccFirewallIPSetCIDRResourceConfig("true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_firewall_ipset_cidr.ten_network", "ipset_name", "proxmoxve_firewall_ipset_test"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_ipset_cidr.ten_network", "cidr", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_ipset_cidr.ten_network", "comment", "open sesame"),
					resource.TestCheckNoResourceAttr("proxmoxve_firewall_ipset_cidr.ten_network", "no_match"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_ipset_cidr.ipv6_ula", "ipset_name", "proxmoxve_firewall_ipset_test"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_ipset_cidr.ipv6_ula", "cidr", "fd65::/16"),
					resource.TestCheckNoResourceAttr("proxmoxve_firewall_ipset_cidr.ipv6_ula", "comment"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_ipset_cidr.ipv6_ula", "no_match", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "proxmoxve_firewall_ipset_cidr.ten_network",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccFirewallIPSetCIDRResourceConfig("false"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_firewall_ipset_cidr.ten_network", "ipset_name", "proxmoxve_firewall_ipset_test"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_ipset_cidr.ten_network", "cidr", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("proxmoxve_firewall_ipset_cidr.ten_network", "comment", "open sesame"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccFirewallIPSetCIDRResourceConfig(v6noMatch string) string {
	return fmt.Sprintf(`
		resource "proxmoxve_firewall_ipset" "test" {
			name    = "proxmoxve_firewall_ipset_test"
			comment = "foobar"
		}

		resource "proxmoxve_firewall_ipset_cidr" "ten_network" {
			ipset_name = proxmoxve_firewall_ipset.test.name
			cidr       = "10.0.0.0/8"
			comment    = "open sesame"
		}

		resource "proxmoxve_firewall_ipset_cidr" "ipv6_ula" {
			ipset_name = proxmoxve_firewall_ipset.test.name
			cidr       = "fd65::/16"
			no_match   = %s
		}
	`, v6noMatch)
}
