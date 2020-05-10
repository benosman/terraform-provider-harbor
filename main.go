package main

import (
	"github.com/benosman/terraform-provider-harbor/harbor"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: harbor.Provider})
}
