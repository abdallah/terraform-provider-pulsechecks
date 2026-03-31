# terraform-provider-pulsechecks

Terraform/OpenTofu provider for [PulseChecks](https://github.com/abdallah/pulsechecks) — a serverless job monitoring service.

## Requirements

- Terraform >= 1.0 or OpenTofu >= 1.6
- Go >= 1.21 (to build from source)

## Usage

```hcl
terraform {
  required_providers {
    pulsechecks = {
      source  = "abdallah/pulsechecks"
      version = "~> 0.1"
    }
  }
}

provider "pulsechecks" {
  # api_token = "..."  # or set PULSECHECKS_API_TOKEN env var
  # api_url   = "..."  # optional: override base URL
}
```

## Authentication

Set the `PULSECHECKS_API_TOKEN` environment variable or pass `api_token` directly in the provider block.

## Resources

### `pulsechecks_check`

```hcl
resource "pulsechecks_check" "daily_backup" {
  team_id        = "my-team-id"
  name           = "Daily Backup"
  check_type     = "heartbeat"   # "heartbeat" or "http"
  period_seconds = 86400
  grace_seconds  = 3600
}

# Use the ping token in your job script:
output "ping_url" {
  value = "https://pulsechecks-api-prod-z5kj4psbnq-uc.a.run.app/ping/${pulsechecks_check.daily_backup.token}"
}
```

HTTP check example:

```hcl
resource "pulsechecks_check" "api_health" {
  team_id             = "my-team-id"
  name                = "API Health"
  check_type          = "http"
  period_seconds      = 300
  grace_seconds       = 60
  url                 = "https://api.example.com/health"
  expected_status_code = 200
  expected_string     = "ok"
  failure_threshold   = 2
}
```

#### Argument Reference

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `team_id` | string | yes | Team ID (forces replacement) |
| `name` | string | yes | Check name |
| `check_type` | string | yes | `heartbeat` or `http` |
| `period_seconds` | int | yes | Expected ping interval |
| `grace_seconds` | int | yes | Grace period before alerting |
| `url` | string | no | URL to monitor (HTTP checks) |
| `expected_status_code` | int | no | Expected HTTP status (default 200) |
| `expected_string` | string | no | Expected response body substring |
| `failure_threshold` | int | no | Failures before alert (default 1) |
| `check_id` | string | computed | Check ID |
| `token` | string | computed | Ping token (sensitive) |

## Data Sources

### `pulsechecks_check`

```hcl
data "pulsechecks_check" "existing" {
  team_id  = "my-team-id"
  check_id = "abc123"
}
```

## Building

```bash
make build    # build binary
make install  # install to ~/.terraform.d/plugins/
make test     # run unit tests
make testacc  # run acceptance tests (requires PULSECHECKS_API_TOKEN + PULSECHECKS_TEAM_ID)
```
