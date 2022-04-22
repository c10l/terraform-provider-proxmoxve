package proxmoxve

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceVersion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccDataSourceVersionConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.proxmoxve_version.test", "version", "7.1-7"),
					resource.TestCheckResourceAttr("data.proxmoxve_version.test", "release", "7.1"),
				),
			},
		},
	})
}

const testAccDataSourceVersionConfig = `data proxmoxve_version test {}`
