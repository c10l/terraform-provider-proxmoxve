package proxmoxve

import (
	"context"

	proxmox "github.com/c10l/proxmoxve-client-go/api2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type versionDatasourceType struct{}

// Version data source schema
func (r versionDatasourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
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
	client *proxmox.Client
}

// NewDataSource -
func (v versionDatasourceType) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return versionDatasource{
		client: p.(*provider).client,
	}, nil
}

// Read -
func (v versionDatasource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data Version

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	version, err := v.client.RetrieveVersion()
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

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
