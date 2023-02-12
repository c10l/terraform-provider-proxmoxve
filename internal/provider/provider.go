package provider

import (
	"context"
	"errors"
	"os"
	"strconv"

	proxmox "github.com/c10l/proxmoxve-client-go/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ProxmoxVEProvider satisfies various provider interfaces.
var _ provider.Provider = &ProxmoxVEProvider{}

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
	TOTPSeed     types.String `tfsdk:"totp_seed"`
	TLSInsecure  types.Bool   `tfsdk:"tls_insecure"`
}

func (p *ProxmoxVEProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "proxmoxve"
	resp.Version = p.version
}

// GetSchema
func (p *ProxmoxVEProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: providerMarkdownDescription,
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Base URL of the Proxmox VE API server. e.g. https://pmve.example.com:8006",
			},
			"token_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "API token ID. e.g. `user@pam!token_name`",
			},
			"secret": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "API Token secret",
			},
			"root_password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Password of the `root` user. Some API endpoints can only be called via a ticket which must be acquired as the `root@pam` user (as opposed to an API token). e.g. the ACME endpoits",
			},
			"totp_seed": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "If the `root` user has 2FA enabled, please inform the seed used to generate the OTPs. At the moment no other methods of 2FA are supported.",
			},
			"tls_insecure": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Set to `true` to bypass TLS cert validation. Defaults to `false`",
			},
		},
	}
}

// Configure -
func (p *ProxmoxVEProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ProxmoxVEProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var baseURL string
	if data.BaseURL.IsUnknown() {
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as URL",
		)
		return
	}
	if data.BaseURL.IsNull() {
		baseURL = os.Getenv("PROXMOXVE_BASE_URL")
	} else {
		baseURL = data.BaseURL.ValueString()
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
	if data.TLSInsecure.IsNull() {
		var err error
		tlsInsecure, err = strconv.ParseBool(os.Getenv("PROXMOXVE_TLS_INSECURE"))
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to parse PROXMOXVE_TLS_INSECURE",
				"PROXMOXVE_TLS_INSECURE needs to be convertible to boolean",
			)
		}
	} else {
		tlsInsecure = data.TLSInsecure.ValueBool()
	}

	clients := map[string]getClientFunc{
		"token": getTokenClientFunc(baseURL, tlsInsecure, data.TokenID, data.Secret),
		"root":  getRootClientFunc(baseURL, tlsInsecure, data.RootPassword, data.TOTPSeed),
	}
	resp.DataSourceData = clients
	resp.ResourceData = clients
}

func getRootClientFunc(baseURL string, insecure bool, rootPassword, totpSeed types.String) func() (*proxmox.Client, error) {
	return func() (*proxmox.Client, error) {
		pwd := rootPassword.ValueString()
		if rootPassword.IsNull() {
			pwd = os.Getenv("PROXMOXVE_ROOT_PASSWORD")
			if pwd == "" {
				return nil, errors.New("root_password cannot be empty")
			}
		}

		totpSd := totpSeed.ValueString()
		if totpSeed.IsNull() {
			totpSd = os.Getenv("PROXMOXVE_TOTPSEED")
		}

		rootClient, err := proxmox.NewTicketClient(baseURL, "root@pam", pwd, totpSd, insecure)
		if err != nil {
			return nil, errors.New("unable to create ProxMox VE client with root@pam user and password:\n\n" + err.Error())
		}

		return rootClient, nil
	}
}

func getTokenClientFunc(baseURL string, insecure bool, tokenID, tokenSecret types.String) func() (*proxmox.Client, error) {
	return func() (*proxmox.Client, error) {
		id := tokenID.ValueString()
		if tokenID.IsNull() {
			id = os.Getenv("PROXMOXVE_TOKEN_ID")
		}
		if id == "" {
			return nil, errors.New("token_id cannot be empty")
		}

		secret := tokenSecret.ValueString()
		if tokenSecret.IsNull() {
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
		NewFirewallIPSetResource,
		NewFirewallIPSetCIDRResource,
		NewFirewallGroupResource,
	}
}

// GetDataSources - Defines provider data sources
func (p *ProxmoxVEProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewVersionDataSource,
		NewStorageDataSource,
		NewFirewallRefsDataSource,
		NewFirewallAliasDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ProxmoxVEProvider{
			version: version,
		}
	}
}
