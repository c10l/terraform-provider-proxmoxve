package proxmoxve

import "github.com/hashicorp/terraform-plugin-framework/types"

type Version struct {
	Release types.String `tfsdk:"release"`
	RepoID  types.String `tfsdk:"repoid"`
	Version types.String `tfsdk:"version"`
}
