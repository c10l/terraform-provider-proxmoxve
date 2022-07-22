package proxmoxve

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestACMEAccountResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccACMEAccountResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_acme_account.test", "id", "terraform_test_account"),
					resource.TestCheckResourceAttr("proxmoxve_acme_account.test", "name", "terraform_test_account"),
					resource.TestCheckResourceAttr("proxmoxve_acme_account.test", "contact", "foo@bar.com"),
					resource.TestCheckResourceAttr("proxmoxve_acme_account.test", "directory", "https://127.0.0.1:14000/dir"),
					resource.TestCheckResourceAttr("proxmoxve_acme_account.test", "tos_url", "foobar"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "proxmoxve_acme_account.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// // Update and Read testing
			// {
			// 	Config: testAccStorageBTRFSResourceConfig([]string{"foo", "baz", "quux"}, `"iso"`),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckTypeSetElemAttr("proxmoxve_acme_account.test", "nodes.*", "foo"),
			// 		resource.TestCheckTypeSetElemAttr("proxmoxve_acme_account.test", "nodes.*", "baz"),
			// 		resource.TestCheckTypeSetElemAttr("proxmoxve_acme_account.test", "nodes.*", "quux"),
			// 		resource.TestCheckTypeSetElemAttr("proxmoxve_acme_account.test", "content.*", "iso"),
			// 	),
			// },
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccACMEAccountResourceConfig() string {
	return `
		resource "proxmoxve_acme_account" "test" {
			name      = "terraform_test_account"
			contact   = "foo@bar.com"
			directory = "https://127.0.0.1:14000/dir"
			tos_url   = "foobar"
		}
	`
}
