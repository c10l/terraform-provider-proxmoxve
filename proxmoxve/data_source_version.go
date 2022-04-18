package proxmoxve

import (
	"context"
	"strconv"
	"time"

	pmapi "github.com/c10l/proxmoxve-client-go/api2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVersion() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVersionRead,
		Schema: map[string]*schema.Schema{
			"release": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repoid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceVersionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	c := m.(*pmapi.Client)

	version, err := c.RetrieveVersion()
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("release", version.Release); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("repoid", version.RepoID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("version", version.Version); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
