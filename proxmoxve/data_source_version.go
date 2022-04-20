package proxmoxve

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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

	v := make(map[string]interface{}, 0)
	err = json.NewDecoder(version).Decode(&v)
	if err != nil {
		return diag.FromErr(err)
	}

	vData, ok := v["data"].(map[string]interface{})
	if !ok {
		var detail []byte
		_, err := version.Read(detail)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could not get version data",
			Detail:   strings.Join([]string{string(detail), err.Error()}, "\n"),
		})
	}

	if err := d.Set("release", vData["release"]); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set release",
			Detail:   fmt.Sprintf("%v", vData),
		})
	}
	if err := d.Set("repoid", vData["repoid"]); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set repoid",
			Detail:   fmt.Sprintf("%v", vData),
		})
	}
	if err := d.Set("version", vData["version"]); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set version",
			Detail:   fmt.Sprintf("%v", vData),
		})
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
