package proxmoxve

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Version struct {
	ID      types.String `tfsdk:"id"`
	Release types.String `tfsdk:"release"`
	RepoID  types.String `tfsdk:"repoid"`
	Version types.String `tfsdk:"version"`
}

type StorageDir struct {
	Storage       types.String `tfsdk:"id"`
	Content       types.Set    `tfsdk:"content"`
	Path          types.String `tfsdk:"path"`
	Type          types.String `tfsdk:"type"`
	Nodes         types.Set    `tfsdk:"nodes"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	PreAllocation types.Bool   `tfsdk:"preallocation"`
}
