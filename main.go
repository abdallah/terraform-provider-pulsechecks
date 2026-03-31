package main

import (
	"context"
	"log"

	"github.com/abdallah/terraform-provider-pulsechecks/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/abdallah/pulsechecks",
	})
	if err != nil {
		log.Fatal(err)
	}
}
