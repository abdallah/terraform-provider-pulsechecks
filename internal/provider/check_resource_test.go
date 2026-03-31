package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/abdallah/terraform-provider-pulsechecks/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"pulsechecks": providerserver.NewProtocol6WithError(provider.New()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("PULSECHECKS_API_TOKEN") == "" {
		t.Skip("PULSECHECKS_API_TOKEN not set, skipping acceptance tests")
	}
	if os.Getenv("PULSECHECKS_TEAM_ID") == "" {
		t.Skip("PULSECHECKS_TEAM_ID not set, skipping acceptance tests")
	}
}

func TestAccCheckResource(t *testing.T) {
	teamID := os.Getenv("PULSECHECKS_TEAM_ID")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckConfig(teamID, "tf-acc-test", 60, 30),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pulsechecks_check.test", "name", "tf-acc-test"),
					resource.TestCheckResourceAttr("pulsechecks_check.test", "period_seconds", "60"),
					resource.TestCheckResourceAttrSet("pulsechecks_check.test", "check_id"),
					resource.TestCheckResourceAttrSet("pulsechecks_check.test", "token"),
				),
			},
			{
				Config: testAccCheckConfig(teamID, "tf-acc-test-updated", 120, 60),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pulsechecks_check.test", "name", "tf-acc-test-updated"),
					resource.TestCheckResourceAttr("pulsechecks_check.test", "period_seconds", "120"),
				),
			},
		},
	})
}

func testAccCheckConfig(teamID, name string, period, grace int) string {
	return fmt.Sprintf(`
provider "pulsechecks" {}

resource "pulsechecks_check" "test" {
  team_id        = %q
  name           = %q
  check_type     = "heartbeat"
  period_seconds = %d
  grace_seconds  = %d
}
`, teamID, name, period, grace)
}
