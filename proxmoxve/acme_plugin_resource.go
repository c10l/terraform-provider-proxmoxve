package proxmoxve

import (
	"context"
	"fmt"
	"os"
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
				Computed: true,
				Type:     types.StringType,
			},
			"name": {
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
	ID types.String `tfsdk:"id"`

	// Required attributes
	Name types.String `tfsdk:"name"`
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

	postReq := plugins.PostRequest{Client: r.provider.rootClient, ID: data.Name.Value, Type: data.Type.Value}
	err := postReq.Post()
	if err != nil {
		resp.Diagnostics.AddError("Error creating acme_plugin", err.Error())
		return
	}

	r.eventuallyGet(ctx, &data, 5*time.Second)

	tflog.Trace(ctx, "created acme_plugin")

	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
}

func (r acmePluginResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data acmePluginResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plugin, err := r.eventuallyGet(ctx, &data, 5*time.Second)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("ACME plugin '%s' does not exist", data.ID.Value)) {
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(fmt.Sprintf("error reading acme_plugin.%s", data.Name.Value), err.Error())
			return
		}
	}

	r.convertAPIGetResponseToTerraform(ctx, *plugin, &data)

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
	tfsdk.ResourceImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r acmePluginResource) convertAPIGetResponseToTerraform(ctx context.Context, apiData plugins.ItemGetResponse, tfData *acmePluginResourceData) {
	tfData.ID = types.String{Value: tfData.Name.Value}
	tfData.Type = types.String{Value: apiData.Type}
}

func (r acmePluginResource) eventuallyGet(ctx context.Context, data *acmePluginResourceData, timeout time.Duration) (*plugins.ItemGetResponse, error) {
	accChan := make(chan *plugins.ItemGetResponse, 1)
	var err error

	go func() {
		var acc *plugins.ItemGetResponse
		elapsedTime := 0 * time.Second
		for {
			acc, err = plugins.ItemGetRequest{Client: r.provider.rootClient, ID: data.Name.Value}.Get()
			if err == nil {
				accChan <- acc
			}
			var wait time.Duration
			if elapsedTime <= 5*time.Second {
				wait = 1 * time.Second
			} else {
				wait = 5 * time.Second
			}
			time.Sleep(wait)
			elapsedTime += wait
			fmt.Fprintf(os.Stderr, "Waiting for proxmoxve_acme_plugins.%s to be created... (%s)\n", data.Name.Value, elapsedTime)
		}
	}()

	select {
	case acc := <-accChan:
		r.convertAPIGetResponseToTerraform(ctx, *acc, data)
		return acc, nil
	case <-time.After(timeout):
		return nil, err
	}
}
