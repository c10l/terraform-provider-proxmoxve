package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	proxmox "github.com/c10l/proxmoxve-client-go/api"
	"github.com/c10l/proxmoxve-client-go/api/cluster/acme/account"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ACMEAccountResource{}
var _ resource.ResourceWithImportState = &ACMEAccountResource{}

func NewACMEAccountResource() resource.Resource {
	return &ACMEAccountResource{}
}

// ACMEAccountResource defines the resource implementation.
type ACMEAccountResource struct {
	client *proxmox.Client
}

// ACMEAccountResource describes the resource data model.
type ACMEAccountResourceModel struct {
	ID types.String `tfsdk:"id"`

	// Required attributes
	Name    types.String `tfsdk:"name"`
	Contact types.String `tfsdk:"contact"`

	// Optional attributes
	Directory types.String `tfsdk:"directory"`
	TOSurl    types.String `tfsdk:"tos_url"`
}

func (r *ACMEAccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acme_account"
}

func (r *ACMEAccountResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
			"contact": {
				Required: true,
				Type:     types.StringType,
			},
			"directory": {
				Optional:      true,
				Type:          types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
			},
			"tos_url": {
				Optional:      true,
				Type:          types.StringType,
				PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
			},
		},
	}, nil
}

func (r *ACMEAccountResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	// TODO: Validate that the `root` password is set. Return an actionable error otherwise.

	client, ok := req.ProviderData.(map[string]*proxmox.Client)["root"]

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *proxmox.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ACMEAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ACMEAccountResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	postReq := account.PostRequest{Client: r.client, Contact: data.Contact.Value}
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

	r.eventuallyGet(ctx, data, 1*time.Second)

	tflog.Trace(ctx, "created resource")

	data.ID = data.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ACMEAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ACMEAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	account, err := r.eventuallyGet(ctx, data, 5*time.Second)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("ACME account config file '%s' does not exist", data.Name.Value)) {
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.AddError(fmt.Sprintf("error reading acme_account.%s", data.Name.Value), err.Error())
			return
		}
	}

	r.convertAPIGetResponseToTerraform(ctx, *account, data)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ACMEAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ACMEAccountResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	putReq := account.ItemPutRequest{Client: r.client, Name: data.Name.Value}
	putReq.Contact = data.Contact.Value
	_, err := putReq.Put()
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error updating acme_account.%s", data.Name.Value), err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *ACMEAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ACMEAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := account.ItemDeleteRequest{Client: r.client, Name: data.Name.Value}.Delete()
	if err != nil {
		resp.Diagnostics.AddError("Error deleting acme_account", err.Error())
		return
	}
}

func (r *ACMEAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *ACMEAccountResource) convertAPIGetResponseToTerraform(ctx context.Context, apiData account.ItemGetResponse, tfData *ACMEAccountResourceModel) {
	tfData.ID = types.String{Value: tfData.Name.Value}
	tfData.Contact = types.String{Value: strings.TrimPrefix(apiData.Account.Contact[0], "mailto:")}
	tfData.Directory = types.String{Value: apiData.Directory}
	tfData.TOSurl = types.String{Value: apiData.TOS}
}

func (r *ACMEAccountResource) eventuallyGet(ctx context.Context, data *ACMEAccountResourceModel, timeout time.Duration) (*account.ItemGetResponse, error) {
	accChan := make(chan *account.ItemGetResponse, 1)
	var err error

	go func() {
		var acc *account.ItemGetResponse
		elapsedTime := 0 * time.Second
		for {
			acc, err = account.ItemGetRequest{Client: r.client, Name: data.Name.Value}.Get()
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
