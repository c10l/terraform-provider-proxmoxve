package provider

import (
	"context"
	"time"

	"github.com/c10l/proxmoxve-client-go/api/cluster/firewall/refs"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.DataSourceType = firewallRefsDatasourceType{}
var _ tfsdk.DataSource = firewallRefsDatasource{}

type firewallRefsDatasourceType struct{}

func (t firewallRefsDatasourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Lists possible IPSet/Alias reference which are allowed in source/dest properties.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"type": {
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "Only list references of specified type. Accepted values: `alias`, `ipset`",
			},
			"refs": {
				Computed: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Computed: true,
					},
					"ref": {
						Type:     types.StringType,
						Computed: true,
					},
					"type": {
						Type:     types.StringType,
						Computed: true,
					},
					"comment": {
						Type:     types.StringType,
						Computed: true,
					},
				}),
			},
		},
	}, nil
}

func (t firewallRefsDatasourceType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return firewallRefsDatasource{
		provider: provider,
	}, diags
}

type firewallRefsDatasourceData struct {
	ID   types.String                    `tfsdk:"id"`
	Type types.String                    `tfsdk:"type"`
	Refs []firewallRefsDatasourceDataRef `tfsdk:"refs"`
}

type firewallRefsDatasourceDataRef struct {
	Name    types.String `tfsdk:"name"`
	Ref     types.String `tfsdk:"ref"`
	Type    types.String `tfsdk:"type"`
	Comment types.String `tfsdk:"comment"`
}

type firewallRefsDatasource struct {
	provider provider
}

func (d firewallRefsDatasource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data firewallRefsDatasourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	firewallRefs, err := refs.GetRequest{Client: d.provider.client}.Do()
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving firewall_refs", err.Error())
		return
	}

	data.ID = types.String{Value: time.Now().String()}
	data.Refs = []firewallRefsDatasourceDataRef{}
	for _, v := range *firewallRefs {
		ref := firewallRefsDatasourceDataRef{}
		ref.Comment = types.String{Value: v.Comment}
		ref.Name = types.String{Value: v.Name}
		ref.Ref = types.String{Value: v.Ref}
		ref.Type = types.String{Value: v.Type}
		data.Refs = append(data.Refs, ref)
	}

	diags = resp.State.Set(ctx, data)
	resp.Diagnostics.Append(diags...)
}
