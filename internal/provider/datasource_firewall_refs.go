package provider

import (
	"context"
	"fmt"
	"time"

	proxmox "github.com/c10l/proxmoxve-client-go/api"
	"github.com/c10l/proxmoxve-client-go/api/cluster/firewall/refs"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &FirewallRefsDataSource{}

// FirewallRefsDataSource -
func NewFirewallRefsDataSource() datasource.DataSource {
	return &FirewallRefsDataSource{}
}

type FirewallRefsDataSource struct {
	client *proxmox.Client
}

type FirewallRefsDataSourceModel struct {
	ID   types.String                         `tfsdk:"id"`
	Type types.String                         `tfsdk:"type"`
	Refs []firewallRefsDataSourceDataRefModel `tfsdk:"refs"`
}

type firewallRefsDataSourceDataRefModel struct {
	Name    types.String `tfsdk:"name"`
	Ref     types.String `tfsdk:"ref"`
	Type    types.String `tfsdk:"type"`
	Comment types.String `tfsdk:"comment"`
}

func (d *FirewallRefsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_refs"
}

func (d *FirewallRefsDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Lists possible IPSet/Alias reference which are allowed in source/dest properties.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"type": {
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "Only list references of specified type. Accepted values: `alias`, `ipset`",
			},
			"refs": {
				Computed: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Computed: true,
					},
					"ref": {
						Type:     types.StringType,
						Computed: true,
					},
					"type": {
						Type:     types.StringType,
						Computed: true,
					},
					"comment": {
						Type:     types.StringType,
						Computed: true,
					},
				}),
			},
		},
	}, nil
}

func (d *FirewallRefsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FirewallRefsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FirewallRefsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	firewallRefs, err := refs.GetRequest{Client: d.client}.Do()
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving firewall_refs", err.Error())
		return
	}

	data.ID = types.String{Value: time.Now().String()}
	data.Refs = []firewallRefsDataSourceDataRefModel{}
	for _, v := range *firewallRefs {
		ref := firewallRefsDataSourceDataRefModel{}
		ref.Comment = types.String{Value: v.Comment}
		ref.Name = types.String{Value: v.Name}
		ref.Ref = types.String{Value: v.Ref}
		ref.Type = types.String{Value: v.Type}
		data.Refs = append(data.Refs, ref)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}
