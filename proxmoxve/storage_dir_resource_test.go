package proxmoxve

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccStorageDirResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccStorageDirResourceConfig("one", "/foo/bar"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_storage_dir.test", "id", "one"),
					resource.TestCheckResourceAttr("proxmoxve_storage_dir.test", "path", "/foo/bar"),
					resource.TestCheckResourceAttr("proxmoxve_storage_dir.test", "disable", "false"),
					resource.TestCheckResourceAttr("proxmoxve_storage_dir.test", "shared", "false"),
					// resource.TestCheckResourceAttr("proxmoxve_storage_dir.test", "type", "dir"),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_dir.test", "content.*", "images"),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_dir.test", "content.*", "rootdir"),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_dir.test", "nodes.*", "foobar"),
				),
			},
			// // ImportState testing
			// {
			// 	ResourceName:      "scaffolding_example.test",
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// 	// This is not normally necessary, but is here because this
			// 	// example code does not have an actual upstream service.
			// 	// Once the Read method is able to refresh information from
			// 	// the upstream service, this can be removed.
			// 	ImportStateVerifyIgnore: []string{"configurable_attribute"},
			// },
			// // Update and Read testing
			// {
			// 	Config: testAccStorageDirResourceConfig("two"),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("scaffolding_example.test", "configurable_attribute", "two"),
			// 	),
			// },
			// // Delete testing automatically occurs in TestCase
		},
	})
}

func testAccStorageDirResourceConfig(storage, path string) string {
	return fmt.Sprintf(`
		resource "proxmoxve_storage_dir" "test" {
			storage = "%[1]s"
			path    = "%[2]s"
			nodes   = ["foobar"]
		}
		`, storage, path)
}
