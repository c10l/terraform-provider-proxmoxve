package proxmoxve

import (
	"context"

	proxmox "github.com/c10l/proxmoxve-client-go/api2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"base_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PROXMOXVE_BASE_URL", nil),
			},
			"token_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PROXMOXVE_TOKEN_ID", nil),
			},
			"secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("PROXMOXVE_SECRET", nil),
			},
			"tls_insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PROXMOXVE_TLS_INSECURE", false),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"proxmoxve_pool": resourcePool(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"proxmoxve_version": dataSourceVersion(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	baseURL := d.Get("base_url").(string)
	tokenID := d.Get("token_id").(string)
	secret := d.Get("secret").(string)
	tlsInsecure := d.Get("tls_insecure").(bool)

	diags := diag.Diagnostics{}

	c, err := proxmox.NewClient(baseURL, tokenID, secret, tlsInsecure)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create ProxMox VE client.",
			Detail:   err.Error(),
		})
	}

	return c, diags
}
