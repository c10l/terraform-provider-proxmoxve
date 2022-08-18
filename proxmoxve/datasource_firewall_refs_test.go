package proxmoxve

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
					resource.TestCheckTypeSetElemAttr("data.proxmoxve_firewall_refs.test", "refs.*", "0"),
					resource.TestCheckNoResourceAttr("data.proxmoxve_firewall_refs.test", "type"),
				),
			},
		},
	})
}

const testAccDataSourceFirewallRefsConfig = `data proxmoxve_firewall_refs test {}`
