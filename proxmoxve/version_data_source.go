package proxmoxve

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type versionDatasourceType struct{}

// Version data source schema
func (r versionDatasourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"release": {
				Type:     types.StringType,
				Computed: true,
			},
			"repoid": {
				Type:     types.StringType,
				Computed: true,
			},
			"version": {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}, nil
}

type versionDatasource struct {
	provider provider
}

// NewDataSource -
func (v versionDatasourceType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return versionDatasource{
		provider: provider,
	}, diags
}

// Read -
func (v versionDatasource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data Version

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	version, err := v.provider.client.RetrieveVersion()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving version",
			err.Error(),
		)
		return
	}

	data.Release = types.String{Value: version.Release}
	data.RepoID = types.String{Value: version.RepoID}
	data.Version = types.String{Value: version.Version}

	data.ID = types.String{Value: version.Version}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
