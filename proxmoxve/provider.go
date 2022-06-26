package proxmoxve

import (
	"context"
	"fmt"
	"os"
	"strconv"

	proxmox "github.com/c10l/proxmoxve-client-go/api"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func New() tfsdk.Provider {
	return &provider{}
}

type provider struct {
	configured bool
	client     *proxmox.Client
}

// GetSchema
func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"base_url": {
				Type:     types.StringType,
				Optional: true,
			},
			"token_id": {
				Type:     types.StringType,
				Optional: true,
			},
			"secret": {
				Type:      types.StringType,
				Optional:  true,
				Sensitive: true,
			},
			"tls_insecure": {
				Type:     types.BoolType,
				Optional: true,
			},
		},
	}, nil
}

// Provider schema struct
type providerData struct {
	BaseURL     types.String `tfsdk:"base_url"`
	TokenID     types.String `tfsdk:"token_id"`
	Secret      types.String `tfsdk:"secret"`
	TLSInsecure types.Bool   `tfsdk:"tls_insecure"`
}

// Configure -
func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	// Retrieve provider data from configuration
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var baseURL string
	if config.BaseURL.Unknown {
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as URL",
		)
		return
	}
	if config.BaseURL.Null {
		baseURL = os.Getenv("PROXMOXVE_BASE_URL")
	} else {
		baseURL = config.BaseURL.Value
	}
	if baseURL == "" {
		// Error vs warning - empty value must stop execution
		resp.Diagnostics.AddError(
			"Unable to find base_url",
			"URL cannot be an empty string",
		)
		return
	}

	var tokenID string
	if config.TokenID.Unknown {
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as Token ID",
		)
		return
	}
	if config.TokenID.Null {
		tokenID = os.Getenv("PROXMOXVE_TOKEN_ID")
	} else {
		tokenID = config.TokenID.Value
	}
	if tokenID == "" {
		// Error vs warning - empty value must stop execution
		resp.Diagnostics.AddError(
			"Unable to find base_url",
			"Token ID cannot be an empty string",
		)
		return
	}

	var secret string
	if config.Secret.Unknown {
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as Secret",
		)
		return
	}
	if config.Secret.Null {
		secret = os.Getenv("PROXMOXVE_SECRET")
	} else {
		secret = config.Secret.Value
	}
	if secret == "" {
		// Error vs warning - empty value must stop execution
		resp.Diagnostics.AddError(
			"Unable to find secret",
			"Secret cannot be an empty string",
		)
		return
	}

	var tlsInsecure bool
	if config.TLSInsecure.Null {
		var err error
		tlsInsecure, err = strconv.ParseBool(os.Getenv("PROXMOXVE_TLS_INSECURE"))
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to parse PROXMOXVE_TLS_INSECURE",
				"PROXMOXVE_TLS_INSECURE needs to be convertible to boolean",
			)
		}
	} else {
		tlsInsecure = config.TLSInsecure.Value
	}

	c, err := proxmox.NewClient(baseURL, tokenID, secret, tlsInsecure)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Unable to create ProxMox VE client:\n\n"+err.Error(),
		)
		return
	}

	p.client = c
	p.configured = true
}

// GetResources - Defines provider resources
func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{}, nil
}

// GetDataSources - Defines provider data sources
func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"proxmoxve_version": versionDatasourceType{},
	}, nil
}

// convertProviderType is a helper function for NewResource and NewDataSource
// implementations to associate the concrete provider type. Alternatively,
// this helper can be skipped and the provider type can be directly type
// asserted (e.g. provider: in.(*provider)), however using this can prevent
// potential panics.
func convertProviderType(in tfsdk.Provider) (provider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*provider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return provider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return provider{}, diags
	}

	return *p, diags
}
