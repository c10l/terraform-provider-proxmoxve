package provider

import (
	"context"
	"os"
	"strconv"

	proxmox "github.com/c10l/proxmoxve-client-go/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ProxmoxVEProvider satisfies various provider interfaces.
var _ provider.Provider = &ProxmoxVEProvider{}
var _ provider.ProviderWithMetadata = &ProxmoxVEProvider{}

type ProxmoxVEProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version    string
	client     *proxmox.Client
	rootClient *proxmox.Client
	// configured bool
}

// Provider schema struct
type ProxmoxVEProviderModel struct {
	BaseURL      types.String `tfsdk:"base_url"`
	TokenID      types.String `tfsdk:"token_id"`
	Secret       types.String `tfsdk:"secret"`
	RootPassword types.String `tfsdk:"root_password"`
	TLSInsecure  types.Bool   `tfsdk:"tls_insecure"`
}

func (p *ProxmoxVEProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "proxmoxve"
	resp.Version = p.version
}

// GetSchema
func (p *ProxmoxVEProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

// Configure -
func (p *ProxmoxVEProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ProxmoxVEProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var baseURL string
	if data.BaseURL.Unknown {
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as URL",
		)
		return
	}
	if data.BaseURL.Null {
		baseURL = os.Getenv("PROXMOXVE_BASE_URL")
	} else {
		baseURL = data.BaseURL.Value
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
	if data.TLSInsecure.Null {
		var err error
		tlsInsecure, err = strconv.ParseBool(os.Getenv("PROXMOXVE_TLS_INSECURE"))
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to parse PROXMOXVE_TLS_INSECURE",
				"PROXMOXVE_TLS_INSECURE needs to be convertible to boolean",
			)
		}
	} else {
		tlsInsecure = data.TLSInsecure.Value
	}

	tokenClient, tokenClientDiags := getTokenClient(baseURL, tlsInsecure, data.TokenID, data.Secret)
	resp.Diagnostics.Append(tokenClientDiags...)
	p.client = tokenClient

	rootClient, rootClientDiags := getRootClient(baseURL, tlsInsecure, data.RootPassword)
	resp.Diagnostics.Append(rootClientDiags...)
	p.rootClient = rootClient

	clients := map[string]*proxmox.Client{
		"token": tokenClient,
		"root":  rootClient,
	}
	resp.DataSourceData = clients
	resp.ResourceData = clients
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
func (p *ProxmoxVEProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// NewStorageDirResource,
		// NewStorageNFSResource,
		// NewStorageBTRFSResource,
		NewACMEAccountResource,
		NewACMEPluginResource,
		// NewFirewallAliasResource,
	}
}

// GetDataSources - Defines provider data sources
func (p *ProxmoxVEProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewVersionDataSource,
		NewStorageDataSource,
		NewFirewallRefsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ProxmoxVEProvider{
			version: version,
		}
	}
}
