package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccStorageBTRFSResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccStorageBTRFSResourceConfig([]string{"foobar"}, `"images","rootdir"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("proxmoxve_storage_btrfs.test", "id", "testacc_storage"),
					resource.TestCheckResourceAttr("proxmoxve_storage_btrfs.test", "path", "/foo/bar"),
					resource.TestCheckResourceAttr("proxmoxve_storage_btrfs.test", "disable", "false"),
					resource.TestCheckResourceAttr("proxmoxve_storage_btrfs.test", "preallocation", ""),
					resource.TestCheckResourceAttr("proxmoxve_storage_btrfs.test", "type", "btrfs"),
					resource.TestCheckResourceAttr("proxmoxve_storage_btrfs.test", "prune_backups", ""),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_btrfs.test", "content.*", "images"),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_btrfs.test", "content.*", "rootdir"),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_btrfs.test", "nodes.*", "foobar"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "proxmoxve_storage_btrfs.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccStorageBTRFSResourceConfig([]string{"foo", "baz", "quux"}, `"iso"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_btrfs.test", "nodes.*", "foo"),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_btrfs.test", "nodes.*", "baz"),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_btrfs.test", "nodes.*", "quux"),
					resource.TestCheckTypeSetElemAttr("proxmoxve_storage_btrfs.test", "content.*", "iso"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccStorageBTRFSResourceConfig(nodes []string, content string) string {
	return fmt.Sprintf(`
		resource "proxmoxve_storage_btrfs" "test" {
			name    = "testacc_storage"
			path    = "/foo/bar"
			nodes   = ["%s"]
			content =	[%s]
		}
		`, strings.Join(nodes, `","`), content)
}
