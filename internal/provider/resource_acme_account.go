package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/c10l/proxmoxve-client-go/api/cluster/acme/account"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
			"name": {
				Required:      true,
				Type:          types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
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
	ID types.String `tfsdk:"id"`

	// Required attributes
	Name    types.String `tfsdk:"name"`
	Contact types.String `tfsdk:"contact"`

	// Optional attributes
	Directory types.String `tfsdk:"directory"`
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

	r.eventuallyGet(ctx, &data, 1*time.Second)

	tflog.Trace(ctx, "created acme_account")

	data.ID = data.Name
	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
}

func (r acmeAccountResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data acmeAccountResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	account, err := r.eventuallyGet(ctx, &data, 5*time.Second)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("ACME account config file '%s' does not exist", data.Name.Value)) {
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(fmt.Sprintf("error reading acme_account.%s", data.Name.Value), err.Error())
			return
		}
	}

	r.convertAPIGetResponseToTerraform(ctx, *account, &data)

	diags = resp.State.Set(ctx, data)
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
		resp.Diagnostics.AddError(fmt.Sprintf("error updating acme_account.%s", data.Name.Value), err.Error())
		return
	}

	diags = resp.State.Set(ctx, data)
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
	tfsdk.ResourceImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r acmeAccountResource) convertAPIGetResponseToTerraform(ctx context.Context, apiData account.ItemGetResponse, tfData *acmeAccountResourceData) {
	tfData.ID = types.String{Value: tfData.Name.Value}
	tfData.Contact = types.String{Value: strings.TrimPrefix(apiData.Account.Contact[0], "mailto:")}
	tfData.Directory = types.String{Value: apiData.Directory}
	tfData.TOSurl = types.String{Value: apiData.TOS}
}

func (r acmeAccountResource) eventuallyGet(ctx context.Context, data *acmeAccountResourceData, timeout time.Duration) (*account.ItemGetResponse, error) {
	accChan := make(chan *account.ItemGetResponse, 1)
	var err error

	go func() {
		var acc *account.ItemGetResponse
		elapsedTime := 0 * time.Second
		for {
			acc, err = account.ItemGetRequest{Client: r.provider.rootClient, Name: data.Name.Value}.Get()
			if err == nil &&
				acc.Directory != "" &&
				acc.Location != "" &&
				acc.TOS != "" &&
				acc.Account.Contact != nil {
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
			fmt.Fprintf(os.Stderr, "Waiting for proxmoxve_acme_account.%s to converge... (%s)\n", data.Name.Value, elapsedTime)
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
