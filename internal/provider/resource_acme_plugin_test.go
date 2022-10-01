package provider

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestACMEPluginResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccACMEPluginResourceConfig("pmve_acme_plugin_test", "foobar", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_acme_plugin.test", "name", "pmve_acme_plugin_test"),
					resource.TestCheckResourceAttr("proxmoxve_acme_plugin.test", "type", "dns"),
					resource.TestCheckResourceAttr("proxmoxve_acme_plugin.test", "data", base64.StdEncoding.EncodeToString([]byte("foobar"))),
					resource.TestCheckTypeSetElemAttr("proxmoxve_acme_plugin.test", "nodes.*", "foobar"),
				),
			},
			{
				Config: testAccACMEPluginResourceConfig("pmve_acme_plugin_test", "foobar", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					// This is supposed to fail
					resource.TestCheckTypeSetElemAttr("proxmoxve_acme_plugin.test", "nodes.*", "foobar"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "proxmoxve_acme_plugin.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccACMEPluginResourceConfig("pmve_acme_plugin_test_update", "bazquux", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_acme_plugin.test", "id", "pmve_acme_plugin_test_update"),
					resource.TestCheckResourceAttr("proxmoxve_acme_plugin.test", "name", "pmve_acme_plugin_test_update"),
					resource.TestCheckResourceAttr("proxmoxve_acme_plugin.test", "data", base64.StdEncoding.EncodeToString([]byte("bazquux"))),
					resource.TestCheckNoResourceAttr("proxmoxve_acme_plugin.test", "nodes"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccACMEPluginResourceConfig(name, data string, addNodes bool) string {
	var nodes string
	if addNodes {
		nodes = `nodes = ["foobar"]`
	} else {
		nodes = ""
	}
	return fmt.Sprintf(`
		resource "proxmoxve_acme_plugin" "test" {
			name  = "%s"
			type  = "dns"
			data  = base64encode("%s")
			api   = "zone"
			%s
		}
	`, name, data, nodes)
}
