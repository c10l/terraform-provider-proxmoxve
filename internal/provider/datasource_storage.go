package provider

import (
	"context"
	"fmt"

	proxmox "github.com/c10l/proxmoxve-client-go/api"
	"github.com/c10l/proxmoxve-client-go/api/storage"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

func (d *StorageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The storage identifier",
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"content": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"path": schema.StringAttribute{
				Computed: true,
			},
			"prune_backups": schema.StringAttribute{
				Computed: true,
			},
			"shared": schema.BoolAttribute{
				Computed: true,
			},
			"nodes": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"enabled": schema.BoolAttribute{
				Computed: true,
			},
			"preallocation": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *StorageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

func (d *StorageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	storage, err := storage.ItemGetRequest{Client: d.client, Storage: data.Name.ValueString()}.Get()
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving version", err.Error())
		return
	}

	var diags diag.Diagnostics
	data.Content, diags = types.SetValueFrom(ctx, types.StringType, storage.Content)
	resp.Diagnostics.Append(diags...)
	data.Nodes, diags = types.SetValueFrom(ctx, types.StringType, storage.Nodes)
	resp.Diagnostics.Append(diags...)

	data.ID = types.StringValue(storage.Storage)
	data.Name = types.StringValue(storage.Storage)
	data.Path = types.StringValue(storage.Path)
	data.PruneBackups = types.StringValue(storage.PruneBackups)
	data.Shared = types.BoolValue(storage.Shared)
	data.Type = types.StringValue(string(storage.Type))

	data.Enabled = types.BoolValue(!storage.Disable)
	data.Preallocation = types.StringValue(storage.Preallocation)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}
