package provider

import (
	"context"
	"fmt"
	"strings"

	proxmox "github.com/c10l/proxmoxve-client-go/api"
	"github.com/c10l/proxmoxve-client-go/api/cluster/firewall/ipset/ipset_cidr"
	"github.com/c10l/proxmoxve-client-go/helpers"
	pvetypes "github.com/c10l/proxmoxve-client-go/helpers/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &FirewallIPSetCIDRResource{}
var _ resource.ResourceWithImportState = &FirewallIPSetCIDRResource{}

func NewFirewallIPSetCIDRResource() resource.Resource {
	return &FirewallIPSetCIDRResource{}
}

// FirewallIPSetCIDRResource defines the resource implementation.
type FirewallIPSetCIDRResource struct {
	client *proxmox.Client
}

// FirewallIPSetCIDRResource describes the resource data model.
type FirewallIPSetCIDRResourceModel struct {
	ID        types.String `tfsdk:"id"`
	IPSetName types.String `tfsdk:"ipset_name"`
	CIDR      types.String `tfsdk:"cidr"`
	NoMatch   types.Bool   `tfsdk:"no_match"`
	Comment   types.String `tfsdk:"comment"`
}

func (r *FirewallIPSetCIDRResource) typeName() string { return "firewall_ipset_cidr" }

func (r *FirewallIPSetCIDRResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.typeName()
}

func (r *FirewallIPSetCIDRResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"ipset_name": schema.StringAttribute{
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "Name of the IPSet on which to attach this CIDR.",
			},
			"cidr": schema.StringAttribute{
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "CIDR to be configured. e.g. `10.0.0.0/8`, `fd65::/16`.",
			},
			"no_match": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Set to `true` to negate the CIDR rather than matching it.",
			},
			"comment": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *FirewallIPSetCIDRResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FirewallIPSetCIDRResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *FirewallIPSetCIDRResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	postReq := ipset_cidr.PostRequest{Client: r.client, IPSetName: data.IPSetName.ValueString(), CIDR: data.CIDR.ValueString()}
	if !data.NoMatch.IsNull() {
		noMatch := pvetypes.PVEBool(data.NoMatch.ValueBool())
		postReq.NoMatch = &noMatch
	}
	if !data.Comment.IsNull() {
		postReq.Comment = helpers.PtrTo(data.Comment.ValueString())
	}
	err := postReq.Post()
	if err != nil {
		resp.Diagnostics.AddError("Error creating "+r.typeName(), err.Error())
		return
	}

	id := fmt.Sprintf("%s/%s", data.IPSetName.ValueString(), data.CIDR.ValueString())
	data.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *FirewallIPSetCIDRResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *FirewallIPSetCIDRResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ipSetCIDR, err := ipset_cidr.ItemGetRequest{Client: r.client, IPSetName: data.IPSetName.ValueString(), CIDR: data.CIDR.ValueString()}.Get()
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading %s %s", r.typeName(), data.ID.ValueString()), err.Error())
		return
	}

	data.CIDR = types.StringValue(ipSetCIDR.CIDR)
	data.Comment = types.StringValue(*ipSetCIDR.Comment)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *FirewallIPSetCIDRResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state *FirewallIPSetCIDRResourceModel
	var config *FirewallIPSetCIDRResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	comment := config.Comment.ValueString()
	noMatch := pvetypes.PVEBool(config.NoMatch.ValueBool())
	itemPutReq := ipset_cidr.ItemPutRequest{
		Client:    r.client,
		IPSetName: config.IPSetName.ValueString(),
		CIDR:      config.CIDR.ValueString(),
		Comment:   &comment,
		NoMatch:   &noMatch,
	}
	err := itemPutReq.Put()
	if err != nil {
		resp.Diagnostics.AddError("Error updating "+r.typeName(), err.Error())
		return
	}

	state.Comment = types.StringValue(comment)
	state.NoMatch = types.BoolValue(bool(noMatch))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *FirewallIPSetCIDRResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *FirewallIPSetCIDRResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := ipset_cidr.ItemDeleteRequest{Client: r.client, IPSetName: data.IPSetName.ValueString(), CIDR: data.CIDR.ValueString()}.Delete()
	if err != nil {
		resp.Diagnostics.AddError("Error deleting "+r.typeName(), err.Error())
		return
	}
}

func (r *FirewallIPSetCIDRResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	ipSetName, cidr := r.splitID(req.ID)
	resp.State.SetAttribute(ctx, path.Root("ipset_name"), ipSetName)
	resp.State.SetAttribute(ctx, path.Root("cidr"), cidr)
}

func (r *FirewallIPSetCIDRResource) splitID(id string) (string, string) {
	nameAndCIDR := strings.SplitN(id, "/", 2)
	return nameAndCIDR[0], nameAndCIDR[1]
}
