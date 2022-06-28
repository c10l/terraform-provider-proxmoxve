package proxmoxve

import (
	"github.com/c10l/proxmoxve-client-go/api/storage"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Version struct {
	ID      types.String `tfsdk:"id"`
	Release types.String `tfsdk:"release"`
	RepoID  types.String `tfsdk:"repoid"`
	Version types.String `tfsdk:"version"`
}

type StorageDir struct {
	Storage       types.String          `tfsdk:"id"`
	Content       storage.ContentList   `tfsdk:"content"`
	Digest        types.String          `tfsdk:"digest,omitempty"`
	Path          types.String          `tfsdk:"path"`
	Type          storage.Type          `tfsdk:"type"`
	Nodes         []types.String        `tfsdk:"nodes"`
	Enabled       types.Bool            `tfsdk:"enabled"`
	PreAllocation storage.Preallocation `tfsdk:"preallocation"`
}
