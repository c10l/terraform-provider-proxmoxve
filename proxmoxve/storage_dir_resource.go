package proxmoxve

import (
	"context"

	"github.com/c10l/proxmoxve-client-go/api/storage"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = storageDirResourceType{}
var _ tfsdk.Resource = storageDirResource{}

// var _ tfsdk.ResourceWithImportState = storageDirResource{}

type storageDirResourceType struct{}

func (t storageDirResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
			// "nodes": {
			// 	Type:     types.StringType,
			// 	Optional: true,
			// },
			// "disable": {
			// 	Type:     types.BoolType,
			// 	Optional: true,
			// },
			// "shared": {
			// 	Type:     types.BoolType,
			// 	Optional: true,
			// },
			// "preallocation": {
			// 	Type:     types.StringType,
			// 	Optional: true,
			// },
			// "type": {
			// 	Type:     types.StringType,
			// 	Computed: true,
			// },
			// "digest": {
			// 	Type:     types.StringType,
			// 	Computed: true,
			// },
			// "prune_backups": {
			// 	Type:     types.BoolType,
			// 	Computed: true,
			// },
		},
	}, nil
}

func (t storageDirResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return storageDirResource{provider: provider}, diags
}

type storageDirResourceData struct {
	Id types.String `tfsdk:"id"`

	// Required attributes
	Storage types.String `tfsdk:"storage"`
	Path    types.String `tfsdk:"path"`

	// Optional attributes
	Content types.Set `tfsdk:"content"`
	// Nodes         types.String `tfsdk:"nodes"`
	// Disable       types.Bool   `tfsdk:"disable"`
	// Shared        types.Bool   `tfsdk:"shared"`
	// Preallocation types.String `tfsdk:"preallocation"`

	// // Computed attributes
	// Type         storage.Type `tfsdk:"type"`
	// Digest       types.String `tfsdk:"digest"`
	// PruneBackups types.String `tfsdk:"prune_backups"`
}

type storageDirResource struct {
	provider provider
}

func (r storageDirResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data storageDirResourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	postReq := storage.PostRequest{Client: r.provider.client, Storage: data.Storage.Value, StorageType: "dir", Path: &data.Path.Value}
	for _, v := range data.Content.Elems {
		value, err := v.ToTerraformValue(ctx)
		if err != nil {
			diagError := diag.NewAttributeErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("content"), "error converting content", err.Error())
			resp.Diagnostics.AddAttributeError(diagError.Path(), diagError.Summary(), diagError.Detail())
		}
		*postReq.Content = append(*postReq.Content, storage.Content(value.String()))
	}
	// if !data.Nodes.Null {
	// 	postReq.Nodes = &data.Nodes.Value
	// }
	// if !data.Disable.Null {
	// 	postReq.Disable = &data.Disable.Value
	// }
	// if !data.Shared.Null {
	// 	postReq.Shared = &data.Shared.Value
	// }
	// if !data.Preallocation.Null {
	// 	v := storage.Preallocation(data.Preallocation.Value)
	// 	postReq.Preallocation = &v
	// }
	_, err := postReq.Do()
	if err != nil {
		resp.Diagnostics.AddError("Error creating storage_dir", err.Error())
		return
	}

	err = r.get(&data)
	if err != nil {
		resp.Diagnostics.AddError("Error reading storage_dir", err.Error())
		return
	}

	tflog.Trace(ctx, "created storage_dir")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r storageDirResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data storageDirResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.get(&data)
	if err != nil {
		resp.Diagnostics.AddError("Error reading storage_dir", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r storageDirResource) get(data *storageDirResourceData) error {
	storage, err := storage.ItemGetRequest{Client: r.provider.client, Storage: data.Storage.Value}.Do()
	if err != nil {
		return err
	}

	data.Content = types.Set{ElemType: types.StringType}
	for _, v := range storage.Content {
		value := types.String{Value: string(v)}
		data.Content.Elems = append(data.Content.Elems, value)
	}

	data.Id = types.String{Value: storage.Storage}
	data.Storage = types.String{Value: storage.Storage}
	data.Path = types.String{Value: storage.Path}
	// data.Digest = types.String{Value: storage.Digest}
	// data.PruneBackups = types.String{Value: storage.PruneBackups}
	// data.Shared = types.Bool{Value: storage.Shared}
	// data.Type = "dir"

	// data.Nodes = types.String{Value: storage.Nodes}
	// data.Disable = types.Bool{Value: storage.Disable}
	// data.Preallocation = types.String{Value: string(storage.Preallocation)}
	return nil
}

func (r storageDirResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

func (r storageDirResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data storageDirResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := storage.ItemDeleteRequest{Client: r.provider.client, Storage: data.Storage.Value}.Do()
	if err != nil {
		resp.Diagnostics.AddError("Error deleting storage_dir", err.Error())
		return
	}
}

func (r storageDirResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
