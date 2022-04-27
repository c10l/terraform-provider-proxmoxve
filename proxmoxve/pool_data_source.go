package proxmoxve

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type poolDatasourceType struct{}

// Pool data source schema
func (r poolDatasourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
				Computed: true,
				Type:     types.StringType,
			},
			"storage_members": {
				Computed: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"content": {
						Computed: true,
						Type:     types.ListType{ElemType: types.StringType},
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

type poolDatasource struct {
	provider provider
}

// NewDataSource -
func (v poolDatasourceType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return poolDatasource{
		provider: provider,
	}, diags
}

// Read -
func (v poolDatasource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var state Pool
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	pool, err := v.provider.client.GetPool(state.PoolID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Failed to retrieve pool", err.Error())
		return
	}

	state.ID = types.String{Value: pool.PoolID}
	state.PoolID = types.String{Value: pool.PoolID}
	state.Comment = types.String{Value: pool.Comment}

	// TODO: Convert pool.Members.Storage into state.StorageMembers
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
