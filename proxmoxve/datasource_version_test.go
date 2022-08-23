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
					resource.TestCheckResourceAttrWith("data.proxmoxve_version.test", "version", testAccRegexpMatch(`7\.\d+-\d+`)),
					resource.TestCheckResourceAttrWith("data.proxmoxve_version.test", "release", testAccRegexpMatch(`7\.\d+`)),
				),
			},
		},
	})
}

const testAccDataSourceVersionConfig = `data proxmoxve_version test {}`
