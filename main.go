package main

import (
	"context"
	"flag"
	"log"
	"terraform-provider-proxmoxve/proxmoxve"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "c10l.cc/local/proxmoxve",
		Debug:   debug,
	}
	err := providerserver.Serve(context.Background(), proxmoxve.New, opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
