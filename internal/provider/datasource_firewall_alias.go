package provider

import (
	"context"
	"fmt"

	proxmox "github.com/c10l/proxmoxve-client-go/api"
	"github.com/c10l/proxmoxve-client-go/api/cluster/firewall/aliases"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &FirewallAliasDataSource{}

// NewFirewallAliasDataSource -
func NewFirewallAliasDataSource() datasource.DataSource {
	return &FirewallAliasDataSource{}
}

type FirewallAliasDataSource struct {
	client *proxmox.Client
}

type FirewallAliasDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	CIDR      types.String `tfsdk:"cidr"`
	IPVersion types.Int64  `tfsdk:"ip_version"`
	Comment   types.String `tfsdk:"comment"`
}

func (d *FirewallAliasDataSource) typeName() string { return "firewall_alias" }

func (d *FirewallAliasDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.typeName()
}

func (d *FirewallAliasDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the firewall alias to retrieve",
			},
			"cidr": schema.StringAttribute{
				Computed: true,
			},
			"comment": schema.StringAttribute{
				Computed: true,
			},
			"ip_version": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *FirewallAliasDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FirewallAliasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FirewallAliasDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alias, err := aliases.ItemGetRequest{Client: d.client, Name: data.Name.ValueString()}.Get()
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading firewall_alias.%s", data.Name.ValueString()), err.Error())
		return
	}

	d.convertAPIGetResponseToTerraform(ctx, *alias, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (d *FirewallAliasDataSource) convertAPIGetResponseToTerraform(ctx context.Context, apiData aliases.ItemGetResponse, tfData *FirewallAliasDataSourceModel) {
	tfData.ID = types.StringValue(apiData.Name)
	tfData.Name = types.StringValue(apiData.Name)
	tfData.CIDR = types.StringValue(apiData.CIDR)
	if apiData.Comment != nil {
		tfData.Comment = types.StringValue(*apiData.Comment)
	}
}
