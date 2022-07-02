package proxmoxve

import (
	"context"
	"fmt"
	"strings"

	"github.com/c10l/proxmoxve-client-go/api/storage"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = storageBTRFSResourceType{}
var _ tfsdk.Resource = storageBTRFSResource{}
var _ tfsdk.ResourceWithImportState = storageBTRFSResource{}

type storageBTRFSResourceType struct{}

func (t storageBTRFSResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Computed: true,
				Type:     types.StringType,
			},
			"storage": {
				Type:     types.StringType,
				Required: true,
			},
			"path": {
				Type:     types.StringType,
				Required: true,
			},
			"content": {
				Type:     types.SetType{ElemType: types.StringType},
				Optional: true,
				Computed: true,
			},
			"nodes": {
				Type:     types.SetType{ElemType: types.StringType},
				Optional: true,
				Computed: true,
			},
			"disable": {
				Type:     types.BoolType,
				Optional: true,
				Computed: true,
			},
			"preallocation": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"type": {
				Type:     types.StringType,
				Computed: true,
			},
			"prune_backups": {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}, nil
}

func (t storageBTRFSResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return storageBTRFSResource{provider: provider}, diags
}

type storageBTRFSResourceData struct {
	Id types.String `tfsdk:"id"`

	// Required attributes
	Storage types.String `tfsdk:"storage"`
	Path    types.String `tfsdk:"path"`

	// Optional attributes
	Content       types.Set    `tfsdk:"content"`
	Nodes         types.Set    `tfsdk:"nodes"`
	Disable       types.Bool   `tfsdk:"disable"`
	Preallocation types.String `tfsdk:"preallocation"`

	// Computed attributes
	Type         types.String `tfsdk:"type"`
	PruneBackups types.String `tfsdk:"prune_backups"`
}

type storageBTRFSResource struct {
	provider provider
}

func (r storageBTRFSResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data storageBTRFSResourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	postReq := storage.PostRequest{Client: r.provider.client, Storage: data.Storage.Value, StorageType: storage.TypeBTRFS, DirPath: &data.Path.Value}
	if !data.Content.Null {
		if postReq.Content == nil {
			postReq.Content = &[]string{}
		}
		resp.Diagnostics.Append(data.Content.ElementsAs(ctx, postReq.Content, false)...)
	}
	if !data.Nodes.Null {
		if postReq.Nodes == nil {
			postReq.Nodes = &[]string{}
		}
		resp.Diagnostics.Append(data.Nodes.ElementsAs(ctx, postReq.Nodes, false)...)
	}
	if !data.Disable.Null {
		postReq.Disable = &data.Disable.Value
	}
	if !data.Preallocation.Null {
		postReq.Preallocation = &data.Preallocation.Value
	}
	_, err := postReq.Do()
	if err != nil {
		resp.Diagnostics.AddError("Error creating storage_btrfs", err.Error())
		return
	}

	storage, err := storage.ItemGetRequest{Client: r.provider.client, Storage: data.Storage.Value}.Do()
	if err != nil {
		resp.Diagnostics.AddError("Error reading storage_btrfs", err.Error())
		return
	}

	resp.Diagnostics.Append(r.convertAPIGetResponseToTerraform(ctx, *storage, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created storage_btrfs")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r storageBTRFSResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data storageBTRFSResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	storage, err := storage.ItemGetRequest{Client: r.provider.client, Storage: data.Storage.Value}.Do()
	if err != nil {
		// If resource has been deleted outside of Terraform, we remove it from the plan state so it can be re-created.
		if strings.Contains(err.Error(), fmt.Sprintf("500 storage '%s' does not exist", data.Storage.Value)) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading storage_btrfs", err.Error())
		return
	}

	resp.Diagnostics.Append(r.convertAPIGetResponseToTerraform(ctx, *storage, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r storageBTRFSResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data storageBTRFSResourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	putReq := storage.ItemPutRequest{Client: r.provider.client, Storage: data.Storage.Value}
	if !data.Content.Null {
		if putReq.Content == nil {
			putReq.Content = &[]string{}
		}
		resp.Diagnostics.Append(data.Content.ElementsAs(ctx, putReq.Content, false)...)
	}
	if !data.Nodes.Null {
		if putReq.Nodes == nil {
			putReq.Nodes = &[]string{}
		}
		resp.Diagnostics.Append(data.Nodes.ElementsAs(ctx, putReq.Nodes, false)...)
	}
	if !data.Disable.Null {
		putReq.Disable = &data.Disable.Value
	}
	if !data.Preallocation.Null {
		putReq.Preallocation = &data.Preallocation.Value
	}
	_, err := putReq.Do()
	if err != nil {
		resp.Diagnostics.AddError("Error creating storage_btrfs", err.Error())
		return
	}

	storage, err := storage.ItemGetRequest{Client: r.provider.client, Storage: data.Storage.Value}.Do()
	if err != nil {
		resp.Diagnostics.AddError("Error reading storage_btrfs", err.Error())
		return
	}

	resp.Diagnostics.Append(r.convertAPIGetResponseToTerraform(ctx, *storage, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated storage_btrfs")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r storageBTRFSResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data storageBTRFSResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := storage.ItemDeleteRequest{Client: r.provider.client, Storage: data.Storage.Value}.Do()
	if err != nil {
		resp.Diagnostics.AddError("Error deleting storage_btrfs", err.Error())
		return
	}
}

func (r storageBTRFSResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("storage"), req, resp)
}

func (r storageBTRFSResource) convertAPIGetResponseToTerraform(ctx context.Context, apiData storage.ItemGetResponse, tfData *storageBTRFSResourceData) diag.Diagnostics {
	var diags diag.Diagnostics
	diags = append(diags, tfsdk.ValueFrom(ctx, apiData.Content, types.SetType{ElemType: types.StringType}, &tfData.Content)...)
	diags = append(diags, tfsdk.ValueFrom(ctx, apiData.Nodes, types.SetType{ElemType: types.StringType}, &tfData.Nodes)...)

	tfData.Id = types.String{Value: apiData.Storage}
	tfData.Storage = types.String{Value: apiData.Storage}
	tfData.Path = types.String{Value: apiData.Path}
	tfData.PruneBackups = types.String{Value: apiData.PruneBackups}
	tfData.Type = types.String{Value: apiData.Type}

	tfData.Disable = types.Bool{Value: apiData.Disable}
	tfData.Preallocation = types.String{Value: apiData.Preallocation}

	return diags
}
