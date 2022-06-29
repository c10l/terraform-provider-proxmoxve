package proxmoxve

import (
	"context"

	"github.com/c10l/proxmoxve-client-go/api/storage"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.DataSourceType = storageDatasourceType{}
var _ tfsdk.DataSource = storageDatasource{}

type storageDatasourceType struct{}

func (t storageDatasourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"type": {
				Type:     types.StringType,
				Computed: true,
			},
			"content": {
				Type:     types.SetType{ElemType: types.StringType},
				Computed: true,
			},
			"path": {
				Type:     types.StringType,
				Computed: true,
			},
			"prune_backups": {
				Type:     types.StringType,
				Computed: true,
			},
			"shared": {
				Type:     types.BoolType,
				Computed: true,
			},
			"storage": {
				Type:     types.StringType,
				Required: true,
			},
			"nodes": {
				Type:     types.SetType{ElemType: types.StringType},
				Computed: true,
			},
			"enabled": {
				Type:     types.BoolType,
				Computed: true,
			},
			"preallocation": {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}, nil
}

func (t storageDatasourceType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return storageDatasource{
		provider: provider,
	}, diags
}

type storageDatasourceData struct {
	Id            types.String `tfsdk:"id"`
	Type          types.String `tfsdk:"type"`
	Content       types.Set    `tfsdk:"content"`
	Path          types.String `tfsdk:"path"`
	PruneBackups  types.String `tfsdk:"prune_backups"`
	Shared        types.Bool   `tfsdk:"shared"`
	Storage       types.String `tfsdk:"storage"`
	Nodes         types.Set    `tfsdk:"nodes"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Preallocation types.String `tfsdk:"preallocation"`
}

type storageDatasource struct {
	provider provider
}

func (d storageDatasource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data storageDatasourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	storage, err := storage.ItemGetRequest{Client: d.provider.client, Storage: data.Storage.Value}.Do()
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving version", err.Error())
		return
	}

	data.Content = types.Set{ElemType: types.StringType}
	for _, v := range storage.Content {
		value := types.String{Value: v}
		data.Content.Elems = append(data.Content.Elems, value)
	}
	data.Nodes = types.Set{ElemType: types.StringType}
	for _, v := range storage.Nodes {
		value := types.String{Value: v}
		data.Content.Elems = append(data.Nodes.Elems, value)
	}

	data.Id = types.String{Value: storage.Storage}
	data.Storage = types.String{Value: storage.Storage}
	data.Path = types.String{Value: storage.Path}
	data.PruneBackups = types.String{Value: storage.PruneBackups}
	data.Shared = types.Bool{Value: storage.Shared}
	data.Type = types.String{Value: string(storage.Type)}

	data.Enabled = types.Bool{Value: !storage.Disable}
	data.Preallocation = types.String{Value: storage.Preallocation}

	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
}
