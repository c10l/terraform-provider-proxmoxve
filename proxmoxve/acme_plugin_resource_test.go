package proxmoxve

import (
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
				Config: testAccACMEPluginResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_acme_plugin.test", "id", "pmve_acme_plugin_test"),
					resource.TestCheckResourceAttr("proxmoxve_acme_plugin.test", "type", "standalone"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "proxmoxve_acme_plugin.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// // Update and Read testing
			// {
			// 	Config: testAccACMEPluginResourceConfig(),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("proxmoxve_acme_plugin.test", "id", "terraform_test_plugin_update"),
			// 		resource.TestCheckResourceAttr("proxmoxve_acme_plugin.test", "name", "terraform_test_plugin_update"),
			// 		resource.TestCheckResourceAttr("proxmoxve_acme_plugin.test", "contact", "foo@barbaz.com"),
			// 		resource.TestCheckResourceAttr("proxmoxve_acme_plugin.test", "directory", "https://127.0.0.1:14000/dir"),
			// 		resource.TestCheckResourceAttr("proxmoxve_acme_plugin.test", "tos_url", "foobar"),
			// 	),
			// },
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccACMEPluginResourceConfig() string {
	return `
		resource "proxmoxve_acme_plugin" "test" {
			id   = "pmve_acme_plugin_test"
			type = "standalone"
		}
	`
}
