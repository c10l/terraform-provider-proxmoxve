package proxmoxve

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type poolResourceType struct{}

// Pool resource schema
func (r poolResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Computed: true,
				Type:     types.StringType,
			},
			"pool_id": {
				Required: true,
				Type:     types.StringType,
			},
			"comment": {
				Optional: true,
				Type:     types.StringType,
			},
			"storage_members": {
				Computed: true,
				Optional: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"content": {
						Computed: true,
						Type:     types.SetType{ElemType: types.StringType},
					},
					"disk": {
						Computed: true,
						Type:     types.Int64Type,
					},
					"id": {
						Computed: true,
						Type:     types.StringType,
					},
					"max_disk": {
						Computed: true,
						Type:     types.Int64Type,
					},
					"node": {
						Computed: true,
						Type:     types.StringType,
					},
					"plugin_type": {
						Computed: true,
						Type:     types.StringType,
					},
					"shared": {
						Computed: true,
						Type:     types.Int64Type,
					},
					"status": {
						Computed: true,
						Type:     types.StringType,
					},
					"storage": {
						Computed: true,
						Type:     types.StringType,
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
		},
	}, nil
}

type poolResource struct {
	provider provider
}

func (t poolResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return poolResource{
		provider: provider,
	}, diags
}

func (r poolResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan Pool
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.provider.client.PostPool(plan.PoolID.Value, plan.Comment.Value); err != nil {
		resp.Diagnostics.AddError("Failed to create pool", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("id"), plan.PoolID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("pool_id"), plan.PoolID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("comment"), plan.Comment)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("storage_members"), []PoolMemberStorage{})...)
}

func (r poolResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state Pool
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	pool, err := r.provider.client.GetPool(state.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Failed to retrieve pool", err.Error())
		return
	}

	state.ID = types.String{Value: pool.PoolID}
	state.PoolID = types.String{Value: pool.PoolID}
	state.Comment = types.String{Value: pool.Comment}

	if len(pool.Members.Storage) > 0 {
		for _, i := range pool.Members.Storage {
			contentList := types.List{}
			for _, i := range i.Content {
				contentList.Elems = append(contentList.Elems, types.String{Value: string(i)})
			}
			state.StorageMembers.Elems = append(state.StorageMembers.Elems, PoolMemberStorage{
				Content:    contentList,
				Disk:       types.Int64{Value: int64(i.Disk)},
				ID:         types.String{Value: i.ID},
				MaxDisk:    types.Int64{Value: int64(i.MaxDisk)},
				Node:       types.String{Value: i.Node},
				PluginType: types.String{Value: string(i.PluginType)},
				Shared:     types.Int64{Value: int64(i.Shared)},
				Status:     types.String{Value: i.Status},
				Storage:    types.String{Value: i.Storage},
			})
		}
		state.StorageMembers.Null = false
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r poolResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan Pool
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.provider.client.PutPool(plan.PoolID.Value, &plan.Comment.Value, nil, nil, false); err != nil {
		resp.Diagnostics.AddError("Failed to create pool", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("comment"), &plan.Comment)...)
}

func (r poolResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state Pool
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.provider.client.DeletePool(state.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete pool", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r poolResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
