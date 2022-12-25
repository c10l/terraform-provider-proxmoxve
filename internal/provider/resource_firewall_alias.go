package provider

import (
	"context"
	"fmt"

	proxmox "github.com/c10l/proxmoxve-client-go/api"
	"github.com/c10l/proxmoxve-client-go/api/cluster/firewall/aliases"
	"github.com/c10l/proxmoxve-client-go/helpers"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &FirewallAliasResource{}
var _ resource.ResourceWithImportState = &FirewallAliasResource{}

func NewFirewallAliasResource() resource.Resource {
	return &FirewallAliasResource{}
}

// FirewallAliasResource defines the resource implementation.
type FirewallAliasResource struct {
	client *proxmox.Client
}

// FirewallAliasResource describes the resource data model.
type FirewallAliasResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	CIDR    types.String `tfsdk:"cidr"`
	Comment types.String `tfsdk:"comment"`
}

func (r *FirewallAliasResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_alias"
}

func (r *FirewallAliasResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"cidr": schema.StringAttribute{
				Required: true,
			},
			"comment": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *FirewallAliasResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *FirewallAliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *FirewallAliasResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	postReq := aliases.PostRequest{Client: r.client, Name: data.Name.ValueString(), CIDR: data.CIDR.ValueString()}
	if !data.Comment.IsNull() {
		postReq.Comment = helpers.PtrTo(data.Comment.ValueString())
	}
	err := postReq.Post()
	if err != nil {
		resp.Diagnostics.AddError("Error creating firewall_alias", err.Error())
		return
	}

	getResp, err := aliases.ItemGetRequest{Client: r.client, Name: data.Name.ValueString()}.Get()
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving firewall_alias", err.Error())
		return
	}
	r.convertAPIGetResponseToTerraform(ctx, *getResp, data)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *FirewallAliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *FirewallAliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alias, err := aliases.ItemGetRequest{Client: r.client, Name: data.Name.ValueString()}.Get()
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading firewall_alias.%s", data.Name.ValueString()), err.Error())
		return
	}

	r.convertAPIGetResponseToTerraform(ctx, *alias, data)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *FirewallAliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var config *FirewallAliasResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	var state *FirewallAliasResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	putReq := aliases.ItemPutRequest{Client: r.client, Name: state.Name.ValueString(), CIDR: config.CIDR.ValueString()}
	if state.Name.ValueString() != config.Name.ValueString() {
		putReq.Rename = helpers.PtrTo(config.Name.ValueString())
	}
	if !config.Comment.IsNull() {
		putReq.Comment = helpers.PtrTo(config.Comment.ValueString())
	}
	err := putReq.Put()
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error updating firewall_alias.%s", state.Name.ValueString()), err.Error())
		return
	}

	getResp, err := aliases.ItemGetRequest{Client: r.client, Name: config.Name.ValueString()}.Get()
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving firewall_alias", err.Error())
		return
	}
	r.convertAPIGetResponseToTerraform(ctx, *getResp, config)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (r *FirewallAliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *FirewallAliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := aliases.ItemDeleteRequest{Client: r.client, Name: data.Name.ValueString()}.Delete()
	if err != nil {
		resp.Diagnostics.AddError("Error deleting firewall_alias", err.Error())
		return
	}
}

func (r *FirewallAliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *FirewallAliasResource) convertAPIGetResponseToTerraform(ctx context.Context, apiData aliases.ItemGetResponse, tfData *FirewallAliasResourceModel) {
	tfData.ID = types.StringValue(apiData.Name)
	tfData.Name = types.StringValue(apiData.Name)
	tfData.CIDR = types.StringValue(apiData.CIDR)
	if apiData.Comment != nil {
		tfData.Comment = types.StringValue(*apiData.Comment)
	}
}
