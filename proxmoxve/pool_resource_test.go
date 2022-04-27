package proxmoxve

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccPoolResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPoolResourceConfig("created"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_pool.test", "pool_id", "test"),
					resource.TestCheckResourceAttr("proxmoxve_pool.test", "comment", "created"),
				),
			},
			// Update and Read testing
			{
				Config: testAccPoolResourceConfig("updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_pool.test", "comment", "updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "proxmoxve_pool.test",
				ImportState:       true,
				ImportStateVerify: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_pool.test", "pool_id", "test"),
					resource.TestCheckResourceAttr("proxmoxve_pool.test", "comment", "updated"),
				),
			},
		},
	})
}

func testAccPoolResourceConfig(comment string) string {
	return fmt.Sprintf(`
resource proxmoxve_pool test {
	pool_id = "test"
	comment = "%s"
}`, comment)
}
