package provider

import (
	"context"
	"fmt"

	proxmox "github.com/c10l/proxmoxve-client-go/api"
	"github.com/c10l/proxmoxve-client-go/api/cluster/firewall/ipset"
	"github.com/c10l/proxmoxve-client-go/helpers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &FirewallIPSetResource{}
var _ resource.ResourceWithImportState = &FirewallIPSetResource{}

func NewFirewallIPSetResource() resource.Resource {
	return &FirewallIPSetResource{}
}

// FirewallIPSetResource defines the resource implementation.
type FirewallIPSetResource struct {
	client *proxmox.Client
}

// FirewallIPSetResource describes the resource data model.
type FirewallIPSetResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Comment types.String `tfsdk:"comment"`
}

func (r *FirewallIPSetResource) typeName() string { return "firewall_ipset" }

func (r *FirewallIPSetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.typeName()
}

func (r *FirewallIPSetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"comment": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *FirewallIPSetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FirewallIPSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *FirewallIPSetResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	postReq := ipset.PostRequest{Client: r.client, Name: data.Name.ValueString()}
	if !data.Comment.IsNull() {
		postReq.Comment = helpers.PtrTo(data.Comment.ValueString())
	}
	err := postReq.Post()
	if err != nil {
		resp.Diagnostics.AddError("Error creating "+r.typeName(), err.Error())
		return
	}

	data.ID = data.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *FirewallIPSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *FirewallIPSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := ipset.ItemGetRequest{Client: r.client, Name: data.Name.ValueString()}.Get()
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading %s %s", r.typeName(), data.Name.ValueString()), err.Error())
		return
	}

	ipSet := r.findIPSetOnList(data.Name.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(ipSet.Name)
	if ipSet.Comment != nil {
		data.Comment = types.StringValue(*ipSet.Comment)
	} else {
		data.Comment = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *FirewallIPSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state *FirewallIPSetResourceModel
	var config *FirewallIPSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rename := state.Name.ValueString()
	var comment *string
	if config.Comment.IsNull() {
		comment = nil
	} else {
		comment = helpers.PtrTo(config.Comment.ValueString())
	}
	putReq := ipset.PostRequest{
		Client:  r.client,
		Name:    config.Name.ValueString(),
		Rename:  &rename,
		Comment: comment,
	}
	err := putReq.Post()
	if err != nil {
		resp.Diagnostics.AddError("Error updating "+r.typeName(), err.Error())
		return
	}

	_, err = ipset.ItemGetRequest{Client: r.client, Name: config.Name.ValueString()}.Get()
	if err != nil {
		resp.Diagnostics.AddError("Error updating "+r.typeName(), err.Error())
		return
	}

	ipSet := r.findIPSetOnList(config.Name.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(ipSet.Name)
	state.Name = types.StringValue(ipSet.Name)
	if ipSet.Comment != nil {
		state.Comment = types.StringValue(*ipSet.Comment)
	} else {
		state.Comment = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *FirewallIPSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *FirewallIPSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := ipset.ItemDeleteRequest{Client: r.client, Name: data.Name.ValueString()}.Delete()
	if err != nil {
		resp.Diagnostics.AddError("Error deleting "+r.typeName(), err.Error())
		return
	}
}

func (r *FirewallIPSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *FirewallIPSetResource) findIPSetOnList(name string, diags *diag.Diagnostics) *ipset.GetResponse {
	ipSetList, err := ipset.GetRequest{Client: r.client}.Get()
	if err != nil {
		diags.AddError("Error getting IPSet list", err.Error())
		return nil
	}

	ipSet := ipSetList.FindByName(name)
	if ipSet == nil {
		diags.AddError(
			fmt.Sprintf("IPSet %s not found on list", name),
			fmt.Sprintf("IPSets returned from the server:\n%+v", ipSetList),
		)
		return nil
	}
	return ipSet

}
