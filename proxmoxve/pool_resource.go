package proxmoxve

import (
	"context"

	proxmox "github.com/c10l/proxmoxve-client-go/api2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type poolResourceType struct{}

// PoolID  string
// Comment string        `json:"comment,omitempty"`
// Members *[]PoolMember `json:"members,omitempty"`

// Pool resource schema
func (r poolResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"pool_id": {
				Type:     types.StringType,
				Computed: true,
			},
			"comment": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"members": {
				Type:     types.ListType{},
				Computed: true,
			},
		},
	}, nil
}

type poolResource struct {
	client *proxmox.Client
}
