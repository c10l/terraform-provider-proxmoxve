package proxmoxve

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type poolDatasourceType struct{}

// Pool data source schema
func (r poolDatasourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return poolResourceType{}.GetSchema(ctx)
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
	var data Pool
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	pool, err := v.provider.client.RetrievePool(data.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving pool", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("id"), &pool.PoolID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("comment"), &pool.Comment)...)
}
