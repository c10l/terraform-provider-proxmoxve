package provider

import (
	"context"
	"fmt"

	proxmox "github.com/c10l/proxmoxve-client-go/api"
	"github.com/c10l/proxmoxve-client-go/api/storage"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &StorageDataSource{}

// NewStorageDataSource -
func NewStorageDataSource() datasource.DataSource {
	return &StorageDataSource{}
}

type StorageDataSource struct {
	client *proxmox.Client
}

type StorageDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Type          types.String `tfsdk:"type"`
	Content       types.Set    `tfsdk:"content"`
	Path          types.String `tfsdk:"path"`
	PruneBackups  types.String `tfsdk:"prune_backups"`
	Shared        types.Bool   `tfsdk:"shared"`
	Nodes         types.Set    `tfsdk:"nodes"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Preallocation types.String `tfsdk:"preallocation"`
}

func (d *StorageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage"
}

func (d *StorageDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"name": {
				Type:                types.StringType,
				Required:            true,
				MarkdownDescription: "The storage identifier",
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

func (d *StorageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(map[string]*proxmox.Client)["token"]

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *proxmox.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *StorageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	storage, err := storage.ItemGetRequest{Client: d.client, Storage: data.Name.Value}.Get()
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

	data.ID = types.String{Value: storage.Storage}
	data.Name = types.String{Value: storage.Storage}
	data.Path = types.String{Value: storage.Path}
	data.PruneBackups = types.String{Value: storage.PruneBackups}
	data.Shared = types.Bool{Value: storage.Shared}
	data.Type = types.String{Value: string(storage.Type)}

	data.Enabled = types.Bool{Value: !storage.Disable}
	data.Preallocation = types.String{Value: storage.Preallocation}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}