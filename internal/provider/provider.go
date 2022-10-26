package provider

import (
	"context"
	"errors"
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

type getClientFunc func() (*proxmox.Client, error)

type ProxmoxVEProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
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
		MarkdownDescription: "Provider configuration. The following environment variables can be set as a fallback for any omitted attributes in the provider declaration: `PROXMOXVE_BASE_URL`, `PROXMOXVE_TOKEN_ID`, `PROXMOXVE_SECRET`, `PROXMOXVE_TLS_INSECURE`, `PROXMOXVE_ROOT_PASSWORD`",
		Attributes: map[string]tfsdk.Attribute{
			"base_url": {
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "Base URL of the Proxmox VE API server. e.g. https://pmve.example.com:8006",
			},
			"token_id": {
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "API token ID. e.g. `user@pam!token_name`",
			},
			"secret": {
				Type:                types.StringType,
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "API Token secret",
			},
			"root_password": {
				Type:                types.StringType,
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Password of the `root` user. Some API endpoints can only be called via a ticket which must be acquired as the `root@pam` user (as opposed to an API token). e.g. the ACME endpoits",
			},
			"tls_insecure": {
				Type:                types.BoolType,
				Optional:            true,
				MarkdownDescription: "Set to `true` to bypass TLS cert validation. Defaults to `false`",
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

	clients := map[string]getClientFunc{
		"token": getTokenClientFunc(baseURL, tlsInsecure, data.TokenID, data.Secret),
		"root":  getRootClientFunc(baseURL, tlsInsecure, data.RootPassword),
	}
	resp.DataSourceData = clients
	resp.ResourceData = clients
}

func getRootClientFunc(baseURL string, insecure bool, rootPassword types.String) func() (*proxmox.Client, error) {
	return func() (*proxmox.Client, error) {
		pwd := rootPassword.Value
		if rootPassword.Null {
			pwd = os.Getenv("PROXMOXVE_ROOT_PASSWORD")
			if pwd == "" {
				return nil, errors.New("root_password cannot be empty")
			}
		}

		rootClient, err := proxmox.NewTicketClient(baseURL, "root@pam", pwd, insecure)
		if err != nil {
			return nil, errors.New("unable to create ProxMox VE client with root@pam user and password:\n\n" + err.Error())
		}

		return rootClient, nil
	}
}

func getTokenClientFunc(baseURL string, insecure bool, tokenID, tokenSecret types.String) func() (*proxmox.Client, error) {
	return func() (*proxmox.Client, error) {
		id := tokenID.Value
		if tokenID.Null {
			id = os.Getenv("PROXMOXVE_TOKEN_ID")
		}
		if id == "" {
			return nil, errors.New("token_id cannot be empty")
		}

		secret := tokenSecret.Value
		if tokenSecret.Null {
			secret = os.Getenv("PROXMOXVE_SECRET")
		}
		if secret == "" {
			return nil, errors.New("secret cannot empty")
		}

		tokenClient, err := proxmox.NewAPITokenClient(baseURL, id, secret, insecure)
		if err != nil {
			return nil, errors.New("unable to create ProxMox VE client with API token:\n\n" + err.Error())
		}

		return tokenClient, nil
	}
}

// GetResources - Defines provider resources
func (p *ProxmoxVEProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewStorageBTRFSResource,
		NewStorageDirResource,
		NewStorageNFSResource,
		NewACMEAccountResource,
		NewACMEPluginResource,
		NewFirewallAliasResource,
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
