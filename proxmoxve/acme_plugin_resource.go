package proxmoxve

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/c10l/proxmoxve-client-go/api/cluster/acme/plugins"
	"github.com/c10l/proxmoxve-client-go/helpers"
	pvetypes "github.com/c10l/proxmoxve-client-go/helpers/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
				Required:      true,
				Type:          types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"type": {
				Required:      true,
				Type:          types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"api": {
				Optional: true,
				Type:     types.StringType,
			},
			"data": {
				Optional: true,
				Type:     types.StringType,
			},
			"disable": {
				Optional: true,
				Computed: true,
				Type:     types.BoolType,
			},
			"nodes": {
				Type:     types.SetType{ElemType: types.StringType},
				Optional: true,
				Computed: true,
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
	API     types.String `tfsdk:"api"`
	Data    types.String `tfsdk:"data"`
	Disable types.Bool   `tfsdk:"disable"`
	Nodes   types.Set    `tfsdk:"nodes"`
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
	if !data.API.Null {
		postReq.API = &data.API.Value
	}
	if !data.Data.Null {
		postReq.Data = &data.Data.Value
	}
	if !data.Disable.Null {
		postReq.Disable = helpers.PtrTo(pvetypes.PVEBool(data.Disable.Value))
	}
	if !data.Nodes.Null {
		postReq.Nodes = &[]string{}
		resp.Diagnostics.Append(data.Nodes.ElementsAs(ctx, postReq.Nodes, false)...)
	}
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
	var config acmePluginResourceData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	putReq := plugins.ItemPutRequest{Client: r.provider.rootClient, ID: config.Name.Value}
	delete := []string{}
	if config.API.Null {
		delete = append(delete, "api")
	} else {
		putReq.API = helpers.PtrTo(config.API.Value)
	}
	if config.Data.Null {
		delete = append(delete, "data")
	} else {
		putReq.Data = helpers.PtrTo(config.Data.Value)
	}
	if config.Disable.Null {
		delete = append(delete, "disable")
	} else {
		putReq.Disable = helpers.PtrTo(pvetypes.PVEBool(config.Disable.Value))
	}
	if config.Nodes.Null {
		delete = append(delete, "nodes")
	} else {
		putReq.Nodes = &[]string{}
		resp.Diagnostics.Append(config.Nodes.ElementsAs(ctx, putReq.Nodes, false)...)
	}
	putReq.Delete = helpers.PtrTo(strings.Join(delete, ","))
	if err := putReq.Put(); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error updating acme_account.%s", config.Name.Value), err.Error())
		return
	}

	state := acmePluginResourceData{}
	plugin, err := r.eventuallyGet(ctx, &state, 5*time.Second)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("ACME plugin '%s' does not exist", state.ID.Value)) {
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(fmt.Sprintf("error reading acme_plugin.%s", state.Name.Value), err.Error())
			return
		}
	}

	r.convertAPIGetResponseToTerraform(ctx, *plugin, &state)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
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
	tfsdk.ResourceImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r acmePluginResource) convertAPIGetResponseToTerraform(ctx context.Context, apiData plugins.ItemGetResponse, tfData *acmePluginResourceData) {
	if apiData.Nodes != nil {
		tfData.Nodes.ElemType = types.StringType
		tfData.Nodes.Elems = []attr.Value{}
		for i, j := range strings.Split(*apiData.Nodes, ",") {
			tfData.Nodes.Elems = append(tfData.Nodes.Elems, types.String{Value: j})
			if i == 0 {
				tfData.Nodes.Null = false
			}
		}
	}

	tfData.ID = types.String{Value: tfData.Name.Value}
	tfData.Type = types.String{Value: apiData.Type}
	if apiData.API != nil {
		tfData.API = types.String{Value: *apiData.API}
	}
	if apiData.Data != nil {
		base64Data := base64.StdEncoding.EncodeToString([]byte(*apiData.Data))
		tfData.Data = types.String{Value: base64Data}
	}
	tfData.Disable = types.Bool{Value: bool(apiData.Disable)}
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
			if elapsedTime < 5*time.Second {
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
