package proxmoxve

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/c10l/proxmoxve-client-go/api/cluster/acme/plugins"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = acmePluginResourceType{}
var _ tfsdk.Resource = acmePluginResource{}
var _ tfsdk.ResourceWithImportState = acmePluginResource{}

type acmePluginResourceType struct{}

func (t acmePluginResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Required: true,
				Type:     types.StringType,
			},
			"type": {
				Required: true,
				Type:     types.StringType,
			},
		},
	}, nil
}

func (t acmePluginResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return acmePluginResource{provider: provider}, diags
}

type acmePluginResourceData struct {
	// Required attributes
	ID   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`

	// Optional attributes
}

type acmePluginResource struct {
	provider provider
}

func (r acmePluginResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data acmePluginResourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	postReq := plugins.PostRequest{Client: r.provider.rootClient, ID: data.ID.Value, Type: data.Type.Value}
	err := postReq.Post()
	if err != nil {
		resp.Diagnostics.AddError("Error creating acme_plugin", err.Error())
		return
	}

	r.getCreatedWithTimeout(ctx, &data, 1*time.Second)

	tflog.Trace(ctx, "created acme_plugin")

	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
}

func (r acmePluginResource) getCreatedWithTimeout(ctx context.Context, data *acmePluginResourceData, timeout time.Duration) (*plugins.ItemGetResponse, diag.Diagnostics) {
	accChan := make(chan *plugins.ItemGetResponse, 1)
	var err error

	go func() {
		var acc *plugins.ItemGetResponse
		elapsedTime := 0 * time.Second
		for {
			acc, err = plugins.ItemGetRequest{Client: r.provider.rootClient, ID: data.ID.Value}.Get()
			if err == nil {
				accChan <- acc
			}
			time.Sleep(5 * time.Second)
			elapsedTime += 5 * time.Second
			fmt.Printf("Waiting for proxmoxve_acme_plugins.%s to be created... (%s)\n", data.ID.Value, elapsedTime)
		}
	}()

	select {
	case acc := <-accChan:
		return acc, r.convertAPIGetResponseToTerraform(ctx, *acc, data)
	case <-time.After(timeout):
		diags := diag.Diagnostics{}
		diags.AddError(fmt.Sprintf("Timed out reading acme_plugins after %s seconds", timeout), err.Error())
		return nil, diags
	}
}

func (r acmePluginResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data acmePluginResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plugin, err := plugins.ItemGetRequest{Client: r.provider.client, ID: data.ID.Value}.Get()
	if err != nil {
		// If resource has been deleted outside of Terraform, we remove it from the plan state so it can be re-created.
		if strings.Contains(err.Error(), fmt.Sprintf("ACME plugin '%s' does not exist", data.ID.Value)) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading acme_plugin", err.Error())
		return
	}

	resp.Diagnostics.Append(r.convertAPIGetResponseToTerraform(ctx, *plugin, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
}

func (r acmePluginResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// TODO: implement
}

func (r acmePluginResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data acmePluginResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := plugins.ItemDeleteRequest{Client: r.provider.rootClient, ID: data.ID.Value}.Delete()
	if err != nil {
		resp.Diagnostics.AddError("Error deleting acme_plugin", err.Error())
		return
	}
}

func (r acmePluginResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// FIXME: Import is not importing `type` argument
	tfsdk.ResourceImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r acmePluginResource) convertAPIGetResponseToTerraform(ctx context.Context, apiData plugins.ItemGetResponse, tfData *acmePluginResourceData) diag.Diagnostics {
	tfData.ID = types.String{Value: tfData.ID.Value}
	tfData.Type = types.String{Value: tfData.Type.Value}

	return nil
}
