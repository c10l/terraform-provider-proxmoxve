package main

import (
	"context"
	"terraform-provider-proxmoxve/proxmoxve"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func main() {
	tfsdk.Serve(context.Background(), proxmoxve.New, tfsdk.ServeOpts{
		Name: "proxmoxve",
	})
}
