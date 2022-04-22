package proxmoxve

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
		resp.Diagnostics.AddError("Error retrieving version", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("id"), &version.Version)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("version"), &version.Version)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("repoid"), &version.RepoID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("release"), &version.Release)...)
}
