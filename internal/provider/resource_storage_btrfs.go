package provider

import (
	"context"
	"fmt"
	"strings"

	proxmox "github.com/c10l/proxmoxve-client-go/api"
	"github.com/c10l/proxmoxve-client-go/api/storage"
	"github.com/c10l/proxmoxve-client-go/helpers"
	pvetypes "github.com/c10l/proxmoxve-client-go/helpers/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &StorageBTRFSResource{}
var _ resource.ResourceWithImportState = &StorageBTRFSResource{}

func NewStorageBTRFSResource() resource.Resource {
	return &StorageBTRFSResource{}
}

// StorageBTRFSResource defines the resource implementation.
type StorageBTRFSResource struct {
	client *proxmox.Client
}

// StorageBTRFSResource describes the resource data model.
type StorageBTRFSResourceModel struct {
	ID types.String `tfsdk:"id"`

	// Required attributes
	Name types.String `tfsdk:"name"`
	Path types.String `tfsdk:"path"`

	// Optional attributes
	Content       types.Set    `tfsdk:"content"`
	Nodes         types.Set    `tfsdk:"nodes"`
	Disable       types.Bool   `tfsdk:"disable"`
	Preallocation types.String `tfsdk:"preallocation"`

	// Computed attributes
	Type         types.String `tfsdk:"type"`
	PruneBackups types.String `tfsdk:"prune_backups"`
}

func (r *StorageBTRFSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_btrfs"
}

func (r *StorageBTRFSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"path": schema.StringAttribute{
				Required: true,
			},
			"content": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"nodes": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"disable": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"preallocation": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"prune_backups": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *StorageBTRFSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StorageBTRFSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *StorageBTRFSResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	postReq := storage.PostRequest{Client: r.client, Storage: data.Name.ValueString(), StorageType: storage.TypeBTRFS, DirPath: helpers.PtrTo(data.Path.ValueString())}
	if !data.Content.IsNull() {
		if postReq.Content == nil {
			postReq.Content = &[]string{}
		}
		resp.Diagnostics.Append(data.Content.ElementsAs(ctx, postReq.Content, false)...)
	}
	if !data.Nodes.IsNull() {
		if postReq.Nodes == nil {
			postReq.Nodes = &[]string{}
		}
		resp.Diagnostics.Append(data.Nodes.ElementsAs(ctx, postReq.Nodes, false)...)
	}
	if !data.Disable.IsNull() {
		postReq.Disable = helpers.PtrTo(pvetypes.PVEBool(data.Disable.ValueBool()))
	}
	if !data.Preallocation.IsNull() {
		postReq.Preallocation = helpers.PtrTo(data.Preallocation.ValueString())
	}
	_, err := postReq.Post()
	if err != nil {
		resp.Diagnostics.AddError("Error creating storage_btrfs", err.Error())
		return
	}

	storage, err := storage.ItemGetRequest{Client: r.client, Storage: data.Name.ValueString()}.Get()
	if err != nil {
		resp.Diagnostics.AddError("Error reading storage_btrfs", err.Error())
		return
	}

	resp.Diagnostics.Append(r.convertAPIGetResponseToTerraform(ctx, *storage, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created storage_btrfs")

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *StorageBTRFSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *StorageBTRFSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	storage, err := storage.ItemGetRequest{Client: r.client, Storage: data.Name.ValueString()}.Get()
	if err != nil {
		// If resource has been deleted outside of Terraform, we remove it from the plan state so it can be re-created.
		if strings.Contains(err.Error(), fmt.Sprintf("500 storage '%s' does not exist", data.Name.ValueString())) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading storage_btrfs", err.Error())
		return
	}

	resp.Diagnostics.Append(r.convertAPIGetResponseToTerraform(ctx, *storage, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *StorageBTRFSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *StorageBTRFSResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	putReq := storage.ItemPutRequest{Client: r.client, Storage: data.Name.ValueString()}
	if !data.Content.IsNull() {
		if putReq.Content == nil {
			putReq.Content = &[]string{}
		}
		resp.Diagnostics.Append(data.Content.ElementsAs(ctx, putReq.Content, false)...)
	}
	if !data.Nodes.IsNull() {
		if putReq.Nodes == nil {
			putReq.Nodes = &[]string{}
		}
		resp.Diagnostics.Append(data.Nodes.ElementsAs(ctx, putReq.Nodes, false)...)
	}
	if !data.Disable.IsNull() {
		putReq.Disable = helpers.PtrTo(pvetypes.PVEBool(data.Disable.ValueBool()))
	}
	if !data.Preallocation.IsNull() {
		putReq.Preallocation = helpers.PtrTo(data.Preallocation.ValueString())
	}
	_, err := putReq.Put()
	if err != nil {
		resp.Diagnostics.AddError("Error creating storage_btrfs", err.Error())
		return
	}

	storage, err := storage.ItemGetRequest{Client: r.client, Storage: data.Name.ValueString()}.Get()
	if err != nil {
		resp.Diagnostics.AddError("Error reading storage_btrfs", err.Error())
		return
	}

	resp.Diagnostics.Append(r.convertAPIGetResponseToTerraform(ctx, *storage, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated storage_btrfs")

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *StorageBTRFSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *StorageBTRFSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := storage.ItemDeleteRequest{Client: r.client, Storage: data.Name.ValueString()}.Delete()
	if err != nil {
		resp.Diagnostics.AddError("Error deleting storage_btrfs", err.Error())
		return
	}
}

func (r *StorageBTRFSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *StorageBTRFSResource) convertAPIGetResponseToTerraform(ctx context.Context, apiData storage.ItemGetResponse, tfData *StorageBTRFSResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags = append(diags, tfsdk.ValueFrom(ctx, apiData.Content, types.SetType{ElemType: types.StringType}, &tfData.Content)...)
	diags = append(diags, tfsdk.ValueFrom(ctx, apiData.Nodes, types.SetType{ElemType: types.StringType}, &tfData.Nodes)...)

	tfData.ID = types.StringValue(apiData.Storage)
	tfData.Name = types.StringValue(apiData.Storage)
	tfData.Path = types.StringValue(apiData.Path)
	tfData.PruneBackups = types.StringValue(apiData.PruneBackups)
	tfData.Type = types.StringValue(apiData.Type)

	tfData.Disable = types.BoolValue(apiData.Disable)
	tfData.Preallocation = types.StringValue(apiData.Preallocation)

	return diags
}
