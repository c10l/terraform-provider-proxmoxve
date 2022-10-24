package provider

import (
	"context"

	version "github.com/c10l/proxmoxve-client-go/api/version"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type versionDatasourceType struct{}

// Version data source schema
func (r versionDatasourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "API version details, including some parts of the global datacenter config.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"release": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "The current Proxmox VE point release in `x.y` format",
			},
			"repoid": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "The short git revision from which this version was build",
			},
			"version": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "The full pve-manager package version of this node",
			},
			"console": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "The default console viewer to use. One of `applet`, `vv`, `html5`, `xtermjs`",
			},
		},
	}, nil
}

type versionDatasource struct {
	provider provider
}

type VersionData struct {
	ID      types.String `tfsdk:"id"`
	Release types.String `tfsdk:"release"`
	RepoID  types.String `tfsdk:"repoid"`
	Version types.String `tfsdk:"version"`
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
	var data VersionData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	version, err := version.GetRequest{Client: v.provider.client}.Do()
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving version", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), version.Version)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("version"), version.Version)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("repoid"), version.RepoID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("release"), version.Release)...)
}
