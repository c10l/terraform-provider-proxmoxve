package provider

import (
	"fmt"
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
				Config: testAccACMEAccountResourceConfig("terraform_test_account", "foo@bar.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
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
			// Update and Read testing
			{
				Config: testAccACMEAccountResourceConfig("terraform_test_account_update", "foo@barbaz.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_acme_account.test", "name", "terraform_test_account_update"),
					resource.TestCheckResourceAttr("proxmoxve_acme_account.test", "contact", "foo@barbaz.com"),
					resource.TestCheckResourceAttr("proxmoxve_acme_account.test", "directory", "https://127.0.0.1:14000/dir"),
					resource.TestCheckResourceAttr("proxmoxve_acme_account.test", "tos_url", "foobar"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccACMEAccountResourceConfig(name, contact string) string {
	return fmt.Sprintf(`
		resource "proxmoxve_acme_account" "test" {
			name      = "%s"
			contact   = "%s"
			directory = "https://127.0.0.1:14000/dir"
			tos_url   = "foobar"
		}
	`, name, contact)
}
