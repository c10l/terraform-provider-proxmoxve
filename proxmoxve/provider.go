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
	rootClient *proxmox.Client
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
			"root_password": {
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
	BaseURL      types.String `tfsdk:"base_url"`
	TokenID      types.String `tfsdk:"token_id"`
	Secret       types.String `tfsdk:"secret"`
	RootPassword types.String `tfsdk:"root_password"`
	TLSInsecure  types.Bool   `tfsdk:"tls_insecure"`
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

	tokenClient, tokenClientDiags := getTokenClient(baseURL, tlsInsecure, config.TokenID, config.Secret)
	resp.Diagnostics.Append(tokenClientDiags...)
	p.client = tokenClient

	rootClient, rootClientDiags := getRootClient(baseURL, tlsInsecure, config.RootPassword)
	resp.Diagnostics.Append(rootClientDiags...)
	p.rootClient = rootClient

	if diags.HasError() {
		return
	}

	p.configured = true
}

func getRootClient(baseURL string, insecure bool, rootPassword types.String) (*proxmox.Client, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	pwd := rootPassword.Value
	if rootPassword.Null {
		pwd = os.Getenv("PROXMOXVE_ROOT_PASSWORD")
	}

	rootClient, err := proxmox.NewTicketClient(baseURL, "root@pam", pwd, insecure)
	if err != nil {
		diags.AddError(
			"Unable to create root@pam ticket client",
			"Unable to create ProxMox VE client with root@pam user and password:\n\n"+err.Error(),
		)
	}

	return rootClient, diags
}

func getTokenClient(baseURL string, insecure bool, tokenID, tokenSecret types.String) (*proxmox.Client, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	id := tokenID.Value
	if tokenID.Null {
		id = os.Getenv("PROXMOXVE_TOKEN_ID")
	}
	if id == "" {
		diags.AddError(
			"Unable to find token_id",
			"Token ID cannot be an empty string",
		)
	}

	secret := tokenSecret.Value
	if tokenSecret.Null {
		secret = os.Getenv("PROXMOXVE_SECRET")
	}
	if secret == "" {
		// Error vs warning - empty value must stop execution
		diags.AddError(
			"Unable to find secret",
			"Secret cannot be an empty string",
		)
	}

	tokenClient, err := proxmox.NewAPITokenClient(baseURL, id, secret, insecure)
	if err != nil {
		diags.AddError(
			"Unable to create token+secret client",
			"Unable to create ProxMox VE client with API token:\n\n"+err.Error(),
		)
	}

	return tokenClient, diags
}

// GetResources - Defines provider resources
func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"proxmoxve_storage_dir":   storageDirResourceType{},
		"proxmoxve_storage_nfs":   storageNFSResourceType{},
		"proxmoxve_storage_btrfs": storageBTRFSResourceType{},
		"proxmoxve_acme_account":  acmeAccountResourceType{},
		"proxmoxve_acme_plugin":   acmePluginResourceType{},
	}, nil
}

// GetDataSources - Defines provider data sources
func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"proxmoxve_version":       versionDatasourceType{},
		"proxmoxve_storage":       storageDatasourceType{},
		"proxmoxve_firewall_refs": firewallRefsDatasourceType{},
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
