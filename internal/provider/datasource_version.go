package provider

import (
	"context"
	"fmt"

	proxmox "github.com/c10l/proxmoxve-client-go/api"
	version "github.com/c10l/proxmoxve-client-go/api/version"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
func (d *VersionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "API version details, including some parts of the global datacenter config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"release": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current Proxmox VE point release in `x.y` format",
			},
			"repoid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The short git revision from which this version was build",
			},
			"version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The full pve-manager package version of this node",
			},
			"console": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The default console viewer to use. One of `applet`, `vv`, `html5`, `xtermjs`",
			},
		},
	}
}

func (d *VersionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	clientFunc, ok := req.ProviderData.(map[string]getClientFunc)["token"]

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *proxmox.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	client, err := clientFunc()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to instantiate client",
			err.Error(),
		)
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
