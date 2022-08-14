package proxmoxve

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceStorage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccDataSourceStoragenConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.proxmoxve_storage.test", "id", "local"),
					resource.TestCheckResourceAttr("data.proxmoxve_storage.test", "name", "local"),
					resource.TestCheckResourceAttr("data.proxmoxve_storage.test", "type", "dir"),
					resource.TestCheckResourceAttr("data.proxmoxve_storage.test", "path", "/var/lib/vz"),
					resource.TestCheckResourceAttr("data.proxmoxve_storage.test", "shared", "false"),
					resource.TestCheckResourceAttr("data.proxmoxve_storage.test", "nodes.#", "0"),
					resource.TestCheckResourceAttr("data.proxmoxve_storage.test", "enabled", "true"),
					resource.TestCheckResourceAttr("data.proxmoxve_storage.test", "preallocation", ""),

					resource.TestCheckTypeSetElemAttr("data.proxmoxve_storage.test", "content.*", "backup"),
					resource.TestCheckTypeSetElemAttr("data.proxmoxve_storage.test", "content.*", "iso"),
					resource.TestCheckTypeSetElemAttr("data.proxmoxve_storage.test", "content.*", "vztmpl"),
				),
			},
		},
	})
}

const testAccDataSourceStoragenConfig = `
data proxmoxve_storage test {
	name = "local"
}`
