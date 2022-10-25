package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	proxmox "github.com/c10l/proxmoxve-client-go/api"
	"github.com/c10l/proxmoxve-client-go/api/cluster/acme/plugins"
	"github.com/c10l/proxmoxve-client-go/helpers"
	pvetypes "github.com/c10l/proxmoxve-client-go/helpers/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ACMEPluginResource{}
var _ resource.ResourceWithImportState = &ACMEPluginResource{}

func NewACMEPluginResource() resource.Resource {
	return &ACMEPluginResource{}
}

// ACMEPluginResource defines the resource implementation.
type ACMEPluginResource struct {
	client *proxmox.Client
}

// ACMEPluginResource describes the resource data model.
type ACMEPluginResourceModel struct {
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

func (r *ACMEPluginResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acme_plugin"
}

func (r *ACMEPluginResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Computed: true,
				Type:     types.StringType,
			},
			"name": {
				Required:      true,
				Type:          types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
			},
			"type": {
				Required:      true,
				Type:          types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
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

func (r *ACMEPluginResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	clientFunc, ok := req.ProviderData.(map[string]getClientFunc)["root"]

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

func (r *ACMEPluginResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ACMEPluginResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	postReq := plugins.PostRequest{Client: r.client, ID: data.Name.Value, Type: data.Type.Value}
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

	r.eventuallyGet(ctx, data, 5*time.Second)

	tflog.Trace(ctx, "created acme_plugin")

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ACMEPluginResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ACMEPluginResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plugin, err := r.eventuallyGet(ctx, data, 5*time.Second)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("ACME plugin '%s' does not exist", data.ID.Value)) {
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(fmt.Sprintf("error reading acme_plugin.%s", data.Name.Value), err.Error())
			return
		}
	}

	r.convertAPIGetResponseToTerraform(ctx, *plugin, data)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ACMEPluginResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ACMEPluginResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	putReq := plugins.ItemPutRequest{Client: r.client, ID: data.Name.Value}
	delete := []string{}
	if data.API.Null {
		delete = append(delete, "api")
	} else {
		putReq.API = helpers.PtrTo(data.API.Value)
	}
	if data.Data.Null {
		delete = append(delete, "data")
	} else {
		putReq.Data = helpers.PtrTo(data.Data.Value)
	}
	if data.Disable.Null {
		delete = append(delete, "disable")
	} else {
		putReq.Disable = helpers.PtrTo(pvetypes.PVEBool(data.Disable.Value))
	}
	if data.Nodes.Null {
		delete = append(delete, "nodes")
	} else {
		putReq.Nodes = &[]string{}
		resp.Diagnostics.Append(data.Nodes.ElementsAs(ctx, putReq.Nodes, false)...)
	}
	putReq.Delete = helpers.PtrTo(strings.Join(delete, ","))
	if err := putReq.Put(); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error updating acme_account.%s", data.Name.Value), err.Error())
		return
	}

	state := ACMEPluginResourceModel{}
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

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ACMEPluginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ACMEPluginResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := plugins.ItemDeleteRequest{Client: r.client, ID: data.ID.Value}.Delete()
	if err != nil {
		resp.Diagnostics.AddError("Error deleting acme_plugin", err.Error())
		return
	}
}

func (r *ACMEPluginResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *ACMEPluginResource) convertAPIGetResponseToTerraform(ctx context.Context, apiData plugins.ItemGetResponse, tfData *ACMEPluginResourceModel) {
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

func (r *ACMEPluginResource) eventuallyGet(ctx context.Context, data *ACMEPluginResourceModel, timeout time.Duration) (*plugins.ItemGetResponse, error) {
	accChan := make(chan *plugins.ItemGetResponse, 1)
	var err error

	go func() {
		var acc *plugins.ItemGetResponse
		elapsedTime := 0 * time.Second
		for {
			acc, err = plugins.ItemGetRequest{Client: r.client, ID: data.Name.Value}.Get()
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
