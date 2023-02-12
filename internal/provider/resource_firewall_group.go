package provider

import (
	"context"
	"fmt"

	proxmox "github.com/c10l/proxmoxve-client-go/api"
	"github.com/c10l/proxmoxve-client-go/api/cluster/firewall/groups"
	"github.com/c10l/proxmoxve-client-go/helpers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &FirewallGroupResource{}
var _ resource.ResourceWithImportState = &FirewallGroupResource{}

func NewFirewallGroupResource() resource.Resource {
	return &FirewallGroupResource{}
}

// FirewallGroupResource defines the resource implementation.
type FirewallGroupResource struct {
	client *proxmox.Client
}

// FirewallGroupResource describes the resource data model.
type FirewallGroupResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Comment types.String `tfsdk:"comment"`
}

func (r *FirewallGroupResource) typeName() string { return "firewall_group" }

func (r *FirewallGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.typeName()
}

func (r *FirewallGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the firewall security group.",
			},
			"comment": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *FirewallGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FirewallGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *FirewallGroupResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	postReq := groups.PostRequest{Client: r.client, Group: data.Name.ValueString()}
	postReq.Comment = helpers.PtrTo(data.Comment.ValueString())
	err := postReq.Post()
	if err != nil {
		resp.Diagnostics.AddError("Error creating "+r.typeName(), err.Error())
		return
	}

	data.ID = data.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *FirewallGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *FirewallGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := r.findGroupOnList(data.Name.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(group.Group)
	data.Name = types.StringValue(group.Group)
	if group.Comment != nil {
		data.Comment = types.StringValue(*group.Comment)
	} else {
		data.Comment = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *FirewallGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state *FirewallGroupResourceModel
	var config *FirewallGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	putReq := groups.PostRequest{
		Client:  r.client,
		Group:   config.Name.ValueString(),
		Rename:  helpers.PtrTo(state.Name.ValueString()),
		Comment: helpers.PtrTo(config.Comment.ValueString()),
	}
	err := putReq.Post()
	if err != nil {
		resp.Diagnostics.AddError("Error updating "+r.typeName(), err.Error())
		return
	}

	config.ID = config.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

func (r *FirewallGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *FirewallGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := groups.ItemDeleteRequest{Client: r.client, Group: data.Name.ValueString()}.Delete()
	if err != nil {
		resp.Diagnostics.AddError("Error deleting "+r.typeName(), err.Error())
		return
	}
}

func (r *FirewallGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *FirewallGroupResource) findGroupOnList(name string, diags *diag.Diagnostics) *groups.GetResponse {
	groupList, err := groups.GetRequest{Client: r.client}.Get()
	if err != nil {
		diags.AddError("Error getting Group list", err.Error())
		return nil
	}

	group := groupList.FindByName(name)
	if group == nil {
		diags.AddError(
			fmt.Sprintf("Group %s not found on list", name),
			fmt.Sprintf("Groups returned from the server:\n%+v", groupList),
		)
		return nil
	}
	return group
}
