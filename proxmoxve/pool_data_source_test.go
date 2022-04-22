package proxmoxve

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccPoolDatasource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPoolDatasourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.proxmoxve_pool.test", "id", "test"),
					resource.TestCheckResourceAttr("data.proxmoxve_pool.test", "comment", "data source read!"),
				),
			},
		},
	})
}

const testAccPoolDatasourceConfig = `
resource proxmoxve_pool test {
	id      = "test"
	comment = "data source read!"
}

data proxmoxve_pool test {
	id         = "test"
	depends_on = [proxmoxve_pool.test]
}`
