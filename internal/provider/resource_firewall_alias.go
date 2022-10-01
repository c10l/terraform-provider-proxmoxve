package provider

import (
	"context"
	"fmt"

	"github.com/c10l/proxmoxve-client-go/api/cluster/firewall/aliases"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = firewallAliasResourceType{}
var _ tfsdk.Resource = firewallAliasResource{}
var _ tfsdk.ResourceWithImportState = firewallAliasResource{}

type firewallAliasResourceType struct{}

func (t firewallAliasResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Computed: true,
				Type:     types.StringType,
			},
			"name": {
				Required: true,
				Type:     types.StringType,
			},
			"cidr": {
				Required: true,
				Type:     types.StringType,
			},
			"digest": {
				Computed: true,
				Type:     types.StringType,
			},
			"comment": {
				Optional: true,
				Type:     types.StringType,
			},
		},
	}, nil
}

func (t firewallAliasResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return firewallAliasResource{provider: provider}, diags
}

type firewallAliasResourceData struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	CIDR    types.String `tfsdk:"cidr"`
	Digest  types.String `tfsdk:"digest"`
	Comment types.String `tfsdk:"comment"`
}

type firewallAliasResource struct {
	provider provider
}

func (r firewallAliasResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data firewallAliasResourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	postReq := aliases.PostRequest{Client: r.provider.client, Name: data.Name.Value, CIDR: data.CIDR.Value}
	if !data.Comment.Null {
		postReq.Comment = &data.Comment.Value
	}
	err := postReq.Post()
	if err != nil {
		resp.Diagnostics.AddError("Error creating firewall_alias", err.Error())
		return
	}

	getResp, err := aliases.ItemGetRequest{Client: r.provider.client, Name: data.Name.Value}.Get()
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving firewall_alias", err.Error())
		return
	}
	r.convertAPIGetResponseToTerraform(ctx, *getResp, &data)

	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
}

func (r firewallAliasResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data firewallAliasResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := aliases.ItemDeleteRequest{Client: r.provider.client, Name: data.Name.Value}.Delete()
	if err != nil {
		resp.Diagnostics.AddError("Error deleting firewall_alias", err.Error())
		return
	}
}

func (r firewallAliasResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data firewallAliasResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	alias, err := aliases.ItemGetRequest{Client: r.provider.client, Name: data.Name.Value}.Get()
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading firewall_alias.%s", data.Name.Value), err.Error())
		return
	}

	// TODO: Read data from the global /firewall/aliases endpoint to get digest

	r.convertAPIGetResponseToTerraform(ctx, *alias, &data)

	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
}

func (r firewallAliasResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var config firewallAliasResourceData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	var state firewallAliasResourceData
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	putReq := aliases.ItemPutRequest{Client: r.provider.client, Name: state.Name.Value, CIDR: config.CIDR.Value}
	if state.Name.Value != config.Name.Value {
		putReq.Rename = &config.Name.Value
	}
	if !config.Comment.Null {
		putReq.Comment = &config.Comment.Value
	}
	err := putReq.Put()
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error updating firewall_alias.%s", state.Name.Value), err.Error())
		return
	}

	getResp, err := aliases.ItemGetRequest{Client: r.provider.client, Name: config.Name.Value}.Get()
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving firewall_alias", err.Error())
		return
	}
	r.convertAPIGetResponseToTerraform(ctx, *getResp, &config)

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

func (r firewallAliasResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r firewallAliasResource) convertAPIGetResponseToTerraform(ctx context.Context, apiData aliases.ItemGetResponse, tfData *firewallAliasResourceData) {
	tfData.ID = types.String{Value: apiData.Name}
	tfData.Name = types.String{Value: apiData.Name}
	tfData.CIDR = types.String{Value: apiData.CIDR}
	if apiData.Comment != nil {
		tfData.Comment = types.String{Value: *apiData.Comment}
	}
}
