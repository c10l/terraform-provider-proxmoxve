package proxmoxve

import (
	"fmt"
	"strings"
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
				Config: testAccStorageDirResourceConfig([]string{"foobar"}, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_storage_dir.test", "id", "testacc_storage"),
					resource.TestCheckResourceAttr("proxmoxve_storage_dir.test", "path", "/foo/bar"),
					resource.TestCheckResourceAttr("proxmoxve_storage_dir.test", "disable", "false"),
					resource.TestCheckResourceAttr("proxmoxve_storage_dir.test", "shared", "false"),
					resource.TestCheckResourceAttr("proxmoxve_storage_dir.test", "preallocation", ""),
					resource.TestCheckResourceAttr("proxmoxve_storage_dir.test", "type", "dir"),
					resource.TestCheckResourceAttr("proxmoxve_storage_dir.test", "prune_backups", ""),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_dir.test", "content.*", "images"),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_dir.test", "content.*", "rootdir"),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_dir.test", "nodes.*", "foobar"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "proxmoxve_storage_dir.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccStorageDirResourceConfig([]string{"foo", "baz", "quux"}, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_dir.test", "nodes.*", "foo"),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_dir.test", "nodes.*", "baz"),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_dir.test", "nodes.*", "quux"),
					resource.TestCheckResourceAttr("proxmoxve_storage_dir.test", "shared", "true"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccStorageDirResourceConfig(nodes []string, shared bool) string {
	return fmt.Sprintf(`
		resource "proxmoxve_storage_dir" "test" {
			name   = "testacc_storage"
			path   = "/foo/bar"
			nodes  = ["%s"]
			shared = %t
		}
		`, strings.Join(nodes, `","`), shared)
}
