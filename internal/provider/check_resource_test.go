package provider_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

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

func teamID() string { return os.Getenv("PULSECHECKS_TEAM_ID") }

// ── Heartbeat ─────────────────────────────────────────────────────────────────

func TestAccCheckResource_Heartbeat(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: heartbeatConfig(teamID(), "acc-heartbeat", 300, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pulsechecks_check.test", "name", "acc-heartbeat"),
					resource.TestCheckResourceAttr("pulsechecks_check.test", "check_type", "heartbeat"),
					resource.TestCheckResourceAttr("pulsechecks_check.test", "period_seconds", "300"),
					resource.TestCheckResourceAttr("pulsechecks_check.test", "grace_seconds", "60"),
					resource.TestCheckResourceAttrSet("pulsechecks_check.test", "check_id"),
					resource.TestCheckResourceAttrSet("pulsechecks_check.test", "token"),
					// schedule must be empty for heartbeat
					resource.TestCheckResourceAttr("pulsechecks_check.test", "schedule", ""),
				),
			},
			// Update name + period — no replacement
			{
				Config: heartbeatConfig(teamID(), "acc-heartbeat-updated", 600, 120),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pulsechecks_check.test", "name", "acc-heartbeat-updated"),
					resource.TestCheckResourceAttr("pulsechecks_check.test", "period_seconds", "600"),
					resource.TestCheckResourceAttr("pulsechecks_check.test", "grace_seconds", "120"),
				),
			},
			// Idempotent plan — no diff after apply
			{
				Config:             heartbeatConfig(teamID(), "acc-heartbeat-updated", 600, 120),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// ── Cron ──────────────────────────────────────────────────────────────────────

func TestAccCheckResource_Cron(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with cron schedule
			{
				Config: cronConfig(teamID(), "acc-cron", "0 2 * * *", 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pulsechecks_check.test", "name", "acc-cron"),
					resource.TestCheckResourceAttr("pulsechecks_check.test", "check_type", "cron"),
					resource.TestCheckResourceAttr("pulsechecks_check.test", "schedule", "0 2 * * *"),
					resource.TestCheckResourceAttrSet("pulsechecks_check.test", "check_id"),
					resource.TestCheckResourceAttrSet("pulsechecks_check.test", "token"),
				),
			},
			// Update schedule
			{
				Config: cronConfig(teamID(), "acc-cron", "0 6 * * 1-5", 30),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pulsechecks_check.test", "schedule", "0 6 * * 1-5"),
					resource.TestCheckResourceAttr("pulsechecks_check.test", "grace_seconds", "30"),
				),
			},
		},
	})
}

// ── HTTP ──────────────────────────────────────────────────────────────────────

func TestAccCheckResource_HTTP(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create HTTP check
			{
				Config: httpConfig(teamID(), "acc-http", "https://wpml.org", 60, 30, 200, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pulsechecks_check.test", "name", "acc-http"),
					resource.TestCheckResourceAttr("pulsechecks_check.test", "check_type", "http"),
					resource.TestCheckResourceAttr("pulsechecks_check.test", "url", "https://wpml.org"),
					resource.TestCheckResourceAttr("pulsechecks_check.test", "expected_status_code", "200"),
					resource.TestCheckResourceAttr("pulsechecks_check.test", "failure_threshold", "2"),
					resource.TestCheckResourceAttr("pulsechecks_check.test", "period_seconds", "60"),
					resource.TestCheckResourceAttrSet("pulsechecks_check.test", "check_id"),
				),
			},
			// Update failure threshold
			{
				Config: httpConfig(teamID(), "acc-http", "https://wpml.org", 60, 30, 200, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pulsechecks_check.test", "failure_threshold", "3"),
				),
			},
		},
	})
}

// ── Data source ───────────────────────────────────────────────────────────────

func TestAccCheckDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: datasourceConfig(teamID()),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Resource attrs
					resource.TestCheckResourceAttr("pulsechecks_check.source", "check_type", "heartbeat"),
					resource.TestCheckResourceAttr("pulsechecks_check.source", "period_seconds", "120"),
					// Data source reads back identical values
					resource.TestCheckResourceAttrPair(
						"data.pulsechecks_check.read", "check_id",
						"pulsechecks_check.source", "check_id",
					),
					resource.TestCheckResourceAttrPair(
						"data.pulsechecks_check.read", "name",
						"pulsechecks_check.source", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.pulsechecks_check.read", "check_type",
						"pulsechecks_check.source", "check_type",
					),
					resource.TestCheckResourceAttrPair(
						"data.pulsechecks_check.read", "period_seconds",
						"pulsechecks_check.source", "period_seconds",
					),
				),
			},
		},
	})
}

// ── Validation: cron without schedule must fail ───────────────────────────────

func TestAccCheckResource_CronWithoutSchedule_Fails(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      cronWithoutScheduleConfig(teamID()),
				ExpectError: regexp.MustCompile(`schedule is required`),
			},
		},
	})
}

// ── Validation: heartbeat without period_seconds must fail ────────────────────

func TestAccCheckResource_HeartbeatWithoutPeriod_Fails(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      heartbeatWithoutPeriodConfig(teamID()),
				ExpectError: regexp.MustCompile(`period_seconds is required`),
			},
		},
	})
}

// ── Config helpers ────────────────────────────────────────────────────────────

func providerBlock() string {
	return `provider "pulsechecks" {}`
}

func heartbeatConfig(tid, name string, period, grace int) string {
	return fmt.Sprintf(`%s
resource "pulsechecks_check" "test" {
  team_id        = %q
  name           = %q
  check_type     = "heartbeat"
  period_seconds = %d
  grace_seconds  = %d
}`, providerBlock(), tid, name, period, grace)
}

func cronConfig(tid, name, schedule string, grace int) string {
	return fmt.Sprintf(`%s
resource "pulsechecks_check" "test" {
  team_id       = %q
  name          = %q
  check_type    = "cron"
  schedule      = %q
  grace_seconds = %d
}`, providerBlock(), tid, name, schedule, grace)
}

func httpConfig(tid, name, url string, period, grace, statusCode, failThreshold int) string {
	return fmt.Sprintf(`%s
resource "pulsechecks_check" "test" {
  team_id              = %q
  name                 = %q
  check_type           = "http"
  period_seconds       = %d
  grace_seconds        = %d
  url                  = %q
  expected_status_code = %d
  failure_threshold    = %d
}`, providerBlock(), tid, name, period, grace, url, statusCode, failThreshold)
}

func datasourceConfig(tid string) string {
	return fmt.Sprintf(`%s
resource "pulsechecks_check" "source" {
  team_id        = %q
  name           = "acc-datasource-test"
  check_type     = "heartbeat"
  period_seconds = 120
  grace_seconds  = 30
}

data "pulsechecks_check" "read" {
  team_id  = %q
  check_id = pulsechecks_check.source.check_id
}`, providerBlock(), tid, tid)
}

func cronWithoutScheduleConfig(tid string) string {
	return fmt.Sprintf(`%s
resource "pulsechecks_check" "test" {
  team_id       = %q
  name          = "acc-invalid-cron"
  check_type    = "cron"
  grace_seconds = 60
}`, providerBlock(), tid)
}

func heartbeatWithoutPeriodConfig(tid string) string {
	return fmt.Sprintf(`%s
resource "pulsechecks_check" "test" {
  team_id       = %q
  name          = "acc-invalid-heartbeat"
  check_type    = "heartbeat"
  grace_seconds = 60
}`, providerBlock(), tid)
}
