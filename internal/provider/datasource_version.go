package provider

import (
	"context"
	"fmt"

	proxmox "github.com/c10l/proxmoxve-client-go/api"
	version "github.com/c10l/proxmoxve-client-go/api/version"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &VersionDataSource{}

// NewVersionDataSource -
func NewVersionDataSource() datasource.DataSource {
	return &VersionDataSource{}
}

type VersionDataSource struct {
	client *proxmox.Client
}

type VersionDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Release types.String `tfsdk:"release"`
	RepoID  types.String `tfsdk:"repoid"`
	Version types.String `tfsdk:"version"`
	Console types.String `tfsdk:"console"`
}

func (d *VersionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_version"
}

// Version data source schema
func (d *VersionDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

func (d *VersionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*proxmox.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *proxmox.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

// Read -
func (d *VersionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VersionDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	version, err := version.GetRequest{Client: d.client}.Do()
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving version", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), version.Version)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("version"), version.Version)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("repoid"), version.RepoID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("release"), version.Release)...)
}
