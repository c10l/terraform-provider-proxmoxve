package proxmoxve

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccStorageNFSResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccStorageNFSResourceConfig("rw", []string{"foobar"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_storage_nfs.test", "id", "testacc_storage_nfs"),
					resource.TestCheckResourceAttr("proxmoxve_storage_nfs.test", "server", "1.2.3.4"),
					resource.TestCheckResourceAttr("proxmoxve_storage_nfs.test", "export", "/mnt/path"),
					resource.TestCheckResourceAttr("proxmoxve_storage_nfs.test", "disable", "true"),
					resource.TestCheckResourceAttr("proxmoxve_storage_nfs.test", "preallocation", ""),
					resource.TestCheckResourceAttr("proxmoxve_storage_nfs.test", "mount_options", "rw"),
					resource.TestCheckResourceAttr("proxmoxve_storage_nfs.test", "type", "nfs"),
					resource.TestCheckResourceAttr("proxmoxve_storage_nfs.test", "prune_backups", ""),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_nfs.test", "content.*", "images"),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_nfs.test", "nodes.*", "foobar"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "proxmoxve_storage_nfs.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccStorageNFSResourceConfig("vers=4.2", []string{"foo", "baz", "quux"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_nfs.test", "nodes.*", "foo"),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_nfs.test", "nodes.*", "baz"),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_nfs.test", "nodes.*", "quux"),
					resource.TestCheckResourceAttr("proxmoxve_storage_nfs.test", "mount_options", "vers=4.2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccStorageNFSResourceConfig(mount_options string, nodes []string) string {
	return fmt.Sprintf(`
		resource "proxmoxve_storage_nfs" "test" {
			name          = "testacc_storage_nfs"
			server        = "1.2.3.4"
			export        = "/mnt/path"
			mount_options = "%s"
			nodes         = ["%s"]
			disable       = true
		}
		`, mount_options, strings.Join(nodes, `","`))
}
