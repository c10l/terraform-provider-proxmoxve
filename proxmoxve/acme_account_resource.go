package proxmoxve

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/c10l/proxmoxve-client-go/api/cluster/acme/account"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = acmeAccountResourceType{}
var _ tfsdk.Resource = acmeAccountResource{}
var _ tfsdk.ResourceWithImportState = acmeAccountResource{}

type acmeAccountResourceType struct{}

func (t acmeAccountResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Computed: true,
				Type:     types.StringType,
			},
			"contact": {
				Required: true,
				Type:     types.StringType,
			},
			"directory": {
				Optional:      true,
				Type:          types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"name": {
				Optional:      true,
				Type:          types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"tos_url": {
				Optional:      true,
				Type:          types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
		},
	}, nil
}

func (t acmeAccountResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return acmeAccountResource{provider: provider}, diags
}

type acmeAccountResourceData struct {
	Id types.String `tfsdk:"id"`

	// Required attributes
	Contact types.String `tfsdk:"contact"`

	// Optional attributes
	Directory types.String `tfsdk:"directory"`
	Name      types.String `tfsdk:"name"`
	TOSurl    types.String `tfsdk:"tos_url"`
}

type acmeAccountResource struct {
	provider provider
}

func (r acmeAccountResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data acmeAccountResourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	postReq := account.PostRequest{Client: r.provider.rootClient, Contact: data.Contact.Value}
	postReq.Name = data.Name.Value
	if !data.Directory.Null {
		postReq.Directory = &data.Directory.Value
	}
	if !data.TOSurl.Null {
		postReq.TOSurl = &data.TOSurl.Value
	}
	_, err := postReq.Post()
	if err != nil {
		resp.Diagnostics.AddError("Error creating acme_account", err.Error())
		return
	}

	r.getCreatedWithTimeout(ctx, &data, 1*time.Second)

	tflog.Trace(ctx, "created acme_account")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r acmeAccountResource) getCreatedWithTimeout(ctx context.Context, data *acmeAccountResourceData, timeout time.Duration) (*account.ItemGetResponse, diag.Diagnostics) {
	accChan := make(chan *account.ItemGetResponse, 1)
	var err error

	go func() {
		var acc *account.ItemGetResponse
		elapsedTime := 0 * time.Second
		for {
			acc, err = account.ItemGetRequest{Client: r.provider.rootClient, Name: data.Name.Value}.Get()
			if err == nil {
				accChan <- acc
			}
			time.Sleep(5 * time.Second)
			elapsedTime += 5 * time.Second
			fmt.Printf("Waiting for proxmoxve_acme_account.%s to be created... (%s)\n", data.Name.Value, elapsedTime)
		}
	}()

	select {
	case acc := <-accChan:
		return acc, r.convertAPIGetResponseToTerraform(ctx, *acc, data)
	case <-time.After(timeout):
		diags := diag.Diagnostics{}
		diags.AddError(fmt.Sprintf("Timed out reading acme_account after %s seconds", timeout), err.Error())
		return nil, diags
	}
}

func (r acmeAccountResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data acmeAccountResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.Value
	account, err := account.ItemGetRequest{Client: r.provider.rootClient, Name: name}.Get()
	if err != nil {
		// If resource has been deleted outside of Terraform, we remove it from the plan state so it can be re-created.
		if strings.Contains(err.Error(), fmt.Sprintf("ACME account config file '%s' does not exist", name)) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading acme_account", err.Error())
		return
	}

	resp.Diagnostics.Append(r.convertAPIGetResponseToTerraform(ctx, *account, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r acmeAccountResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data acmeAccountResourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	putReq := account.ItemPutRequest{Client: r.provider.rootClient, Name: data.Name.Value}
	putReq.Contact = data.Contact.Value
	_, err := putReq.Put()
	if err != nil {
		resp.Diagnostics.AddError("Error updating acme_account", err.Error())
		return
	}

	account, err := account.ItemGetRequest{Client: r.provider.client, Name: data.Name.Value}.Get()
	if err != nil {
		resp.Diagnostics.AddError("Error reading acme_account", err.Error())
		return
	}

	resp.Diagnostics.Append(r.convertAPIGetResponseToTerraform(ctx, *account, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.convertAPIGetResponseToTerraform(ctx, *account, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r acmeAccountResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data acmeAccountResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := account.ItemDeleteRequest{Client: r.provider.rootClient, Name: data.Name.Value}.Delete()
	if err != nil {
		resp.Diagnostics.AddError("Error deleting acme_account", err.Error())
		return
	}
}

func (r acmeAccountResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("name"), req, resp)
}

func (r acmeAccountResource) convertAPIGetResponseToTerraform(ctx context.Context, apiData account.ItemGetResponse, tfData *acmeAccountResourceData) diag.Diagnostics {
	tfData.Contact = types.String{Value: strings.TrimPrefix(apiData.Account.Contact[0], "mailto:")}
	tfData.Directory = types.String{Value: apiData.Directory}
	tfData.TOSurl = types.String{Value: apiData.TOS}
	tfData.Id = types.String{Value: tfData.Name.Value}

	return nil
}
