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
var _ resource.Resource = &StorageNFSResource{}
var _ resource.ResourceWithImportState = &StorageNFSResource{}

func NewStorageNFSResource() resource.Resource {
	return &StorageNFSResource{}
}

// StorageNFSResource defines the resource implementation.
type StorageNFSResource struct {
	client *proxmox.Client
}

// StorageNFSResource describes the resource data model.
type StorageNFSResourceModel struct {
	ID types.String `tfsdk:"id"`

	// Required attributes
	Name   types.String `tfsdk:"name"`
	Server types.String `tfsdk:"server"`
	Export types.String `tfsdk:"export"`

	// Optional attributes
	Content       types.Set    `tfsdk:"content"`
	Nodes         types.Set    `tfsdk:"nodes"`
	Disable       types.Bool   `tfsdk:"disable"`
	Preallocation types.String `tfsdk:"preallocation"`
	MountOptions  types.String `tfsdk:"mount_options"`

	// Computed attributes
	Type         types.String `tfsdk:"type"`
	PruneBackups types.String `tfsdk:"prune_backups"`
}

func (r *StorageNFSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_nfs"
}

func (r *StorageNFSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"server": schema.StringAttribute{
				Required: true,
			},
			"export": schema.StringAttribute{
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
			"mount_options": schema.StringAttribute{
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

func (r *StorageNFSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StorageNFSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *StorageNFSResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	postReq := storage.PostRequest{
		Client:      r.client,
		Storage:     data.Name.ValueString(),
		StorageType: storage.TypeNFS,
		NFSServer:   helpers.PtrTo(data.Server.ValueString()),
		NFSExport:   helpers.PtrTo(data.Export.ValueString()),
	}
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
	if !data.MountOptions.IsNull() {
		postReq.NFSMountOptions = helpers.PtrTo(data.MountOptions.ValueString())
	}
	if !data.Preallocation.IsNull() {
		postReq.Preallocation = helpers.PtrTo(data.Preallocation.ValueString())
	}
	_, err := postReq.Post()
	if err != nil {
		resp.Diagnostics.AddError("Error creating storage_nfs", err.Error())
		return
	}

	storage, err := storage.ItemGetRequest{Client: r.client, Storage: data.Name.ValueString()}.Get()
	if err != nil {
		resp.Diagnostics.AddError("Error reading storage_nfs", err.Error())
		return
	}

	resp.Diagnostics.Append(r.convertAPIGetResponseToTerraform(ctx, *storage, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created storage_nfs")

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *StorageNFSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *StorageNFSResourceModel
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
		resp.Diagnostics.AddError("Error reading storage_nfs", err.Error())
		return
	}

	resp.Diagnostics.Append(r.convertAPIGetResponseToTerraform(ctx, *storage, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *StorageNFSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *StorageNFSResourceModel
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
	if !data.MountOptions.IsNull() {
		putReq.NFSMountOptions = helpers.PtrTo(data.MountOptions.ValueString())
	}
	if !data.Preallocation.IsNull() {
		putReq.Preallocation = helpers.PtrTo(data.Preallocation.ValueString())
	}
	_, err := putReq.Put()
	if err != nil {
		resp.Diagnostics.AddError("Error creating storage_nfs", err.Error())
		return
	}

	storage, err := storage.ItemGetRequest{Client: r.client, Storage: data.Name.ValueString()}.Get()
	if err != nil {
		resp.Diagnostics.AddError("Error reading storage_nfs", err.Error())
		return
	}

	resp.Diagnostics.Append(r.convertAPIGetResponseToTerraform(ctx, *storage, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated storage_nfs")

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *StorageNFSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *StorageNFSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := storage.ItemDeleteRequest{Client: r.client, Storage: data.Name.ValueString()}.Delete()
	if err != nil {
		resp.Diagnostics.AddError("Error deleting storage_nfs", err.Error())
		return
	}
}

func (r *StorageNFSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *StorageNFSResource) convertAPIGetResponseToTerraform(ctx context.Context, apiData storage.ItemGetResponse, tfData *StorageNFSResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags = append(diags, tfsdk.ValueFrom(ctx, apiData.Content, types.SetType{ElemType: types.StringType}, &tfData.Content)...)
	diags = append(diags, tfsdk.ValueFrom(ctx, apiData.Nodes, types.SetType{ElemType: types.StringType}, &tfData.Nodes)...)

	tfData.ID = types.StringValue(apiData.Storage)
	tfData.Name = types.StringValue(apiData.Storage)
	tfData.Server = types.StringValue(apiData.NFSServer)
	tfData.PruneBackups = types.StringValue(apiData.PruneBackups)
	tfData.MountOptions = types.StringValue(apiData.NFSMountOptions)
	tfData.Export = types.StringValue(apiData.NFSExport)
	tfData.Type = types.StringValue(apiData.Type)

	tfData.Disable = types.BoolValue(apiData.Disable)
	tfData.Preallocation = types.StringValue(apiData.Preallocation)

	return diags
}
