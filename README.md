# terraform-provider-pulsechecks

[![Terraform Registry](https://img.shields.io/badge/Terraform%20Registry-abdallah%2Fpulsechecks-blue?logo=terraform)](https://registry.terraform.io/providers/abdallah/pulsechecks/latest)
[![Release](https://img.shields.io/github/v/release/abdallah/terraform-provider-pulsechecks)](https://github.com/abdallah/terraform-provider-pulsechecks/releases)
[![License: MPL-2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](LICENSE)

Terraform / OpenTofu provider for [PulseChecks](https://github.com/abdallah/pulsechecks) — a lightweight, serverless uptime and cron-job monitoring service.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0 **or** [OpenTofu](https://opentofu.org/docs/intro/install/) >= 1.6
- Go >= 1.21 (only if building from source)

## Installation

```hcl
terraform {
  required_providers {
    pulsechecks = {
      source  = "abdallah/pulsechecks"
      version = "~> 0.1"
    }
  }
}
```

Then run:

```bash
terraform init
```

## Authentication

Create an API token in the PulseChecks team settings page, then provide it to the provider via environment variable (recommended) or inline:

**Environment variable (recommended)**

```bash
export PULSECHECKS_API_TOKEN="pc_..."
export PULSECHECKS_TEAM_ID="your-team-id"
export PULSECHECKS_API_URL="https://pulsechecks-api-prod-z5kj4psbnq-uc.a.run.app"  # optional
```

**Provider block**

```hcl
provider "pulsechecks" {
  api_token = "pc_..."                                                          # or PULSECHECKS_API_TOKEN
  team_id   = "your-team-id"                                                    # or PULSECHECKS_TEAM_ID
  api_url   = "https://pulsechecks-api-prod-z5kj4psbnq-uc.a.run.app"          # or PULSECHECKS_API_URL
}
```

| Argument    | Env var                  | Description                              |
|-------------|--------------------------|------------------------------------------|
| `api_token` | `PULSECHECKS_API_TOKEN`  | API token (`pc_` prefix) — **required**  |
| `team_id`   | `PULSECHECKS_TEAM_ID`    | Team ID — **required**                   |
| `api_url`   | `PULSECHECKS_API_URL`    | API base URL (defaults to hosted service)|

## Resources

### `pulsechecks_check`

Manages a PulseChecks health check. Three check types are supported:

#### Heartbeat check

A job pings the check URL on each successful run. Alert fires if no ping is received within `period_seconds + grace_seconds`.

```hcl
resource "pulsechecks_check" "daily_backup" {
  name           = "Daily Backup"
  check_type     = "heartbeat"
  period_seconds = 86400   # expect a ping every 24h
  grace_seconds  = 3600    # alert after 1h of silence
}

# Ping URL for your job script:
output "ping_url" {
  value = "https://pulsechecks-api-prod-z5kj4psbnq-uc.a.run.app/ping/${pulsechecks_check.daily_backup.token}"
}
```

#### HTTP check

PulseChecks polls the URL at regular intervals and alerts on unexpected status codes or missing response strings.

```hcl
resource "pulsechecks_check" "api_health" {
  name                 = "API Health"
  check_type           = "http"
  period_seconds       = 300     # poll every 5 minutes
  grace_seconds        = 60
  url                  = "https://api.example.com/health"
  expected_status_code = 200
  expected_string      = "ok"
  failure_threshold    = 2       # alert after 2 consecutive failures
}
```

#### Cron check

Monitors a scheduled job by schedule expression. Alerts if the job doesn't ping within the grace period after the expected run time.

```hcl
resource "pulsechecks_check" "nightly_report" {
  name          = "Nightly Report"
  check_type    = "cron"
  schedule      = "0 2 * * *"   # every day at 02:00 UTC
  grace_seconds = 1800          # 30-minute grace window
}
```

#### Argument reference

| Argument              | Type   | Required               | Description                                    |
|-----------------------|--------|------------------------|------------------------------------------------|
| `name`                | string | yes                    | Display name                                   |
| `check_type`          | string | yes                    | `heartbeat`, `http`, or `cron`                 |
| `period_seconds`      | int    | heartbeat / http only  | Expected ping interval in seconds              |
| `grace_seconds`       | int    | yes                    | Grace period before alerting (seconds)         |
| `schedule`            | string | cron only              | Cron expression (e.g. `"0 2 * * *"`)          |
| `url`                 | string | http only              | URL to poll                                    |
| `expected_status_code`| int    | no                     | Expected HTTP status code (default `200`)      |
| `expected_string`     | string | no                     | Expected substring in response body            |
| `failure_threshold`   | int    | no                     | Consecutive failures before alert (default `1`)|

#### Computed attributes

| Attribute    | Description                           |
|--------------|---------------------------------------|
| `check_id`   | Unique check ID                       |
| `token`      | Ping token (sensitive) — use as `${resource.token}` |
| `status`     | Current check status                  |
| `created_at` | Creation timestamp                    |

## Data Sources

### `pulsechecks_check`

Look up an existing check by ID.

```hcl
data "pulsechecks_check" "existing" {
  team_id  = "your-team-id"
  check_id = "abc123"
}

output "check_status" {
  value = data.pulsechecks_check.existing.status
}
```

## Complete Example

```hcl
terraform {
  required_providers {
    pulsechecks = {
      source  = "abdallah/pulsechecks"
      version = "~> 0.1"
    }
  }
}

provider "pulsechecks" {}  # reads from PULSECHECKS_API_TOKEN / PULSECHECKS_TEAM_ID

resource "pulsechecks_check" "db_backup" {
  name           = "Database Backup"
  check_type     = "heartbeat"
  period_seconds = 86400
  grace_seconds  = 3600
}

resource "pulsechecks_check" "web" {
  name                 = "Website"
  check_type           = "http"
  period_seconds       = 60
  grace_seconds        = 30
  url                  = "https://example.com"
  expected_status_code = 200
  failure_threshold    = 2
}

resource "pulsechecks_check" "weekly_report" {
  name          = "Weekly Report"
  check_type    = "cron"
  schedule      = "0 9 * * 1"  # every Monday at 09:00 UTC
  grace_seconds = 3600
}

output "db_backup_ping_url" {
  value     = "https://pulsechecks-api-prod-z5kj4psbnq-uc.a.run.app/ping/${pulsechecks_check.db_backup.token}"
  sensitive = true
}
```

## Development

### Building from source

```bash
git clone https://github.com/abdallah/terraform-provider-pulsechecks
cd terraform-provider-pulsechecks
go build ./...
```

### Testing

```bash
# Unit tests
go test ./internal/provider/... -v

# Acceptance tests (requires a live PulseChecks API)
export PULSECHECKS_API_TOKEN="pc_..."
export PULSECHECKS_TEAM_ID="your-team-id"
export PULSECHECKS_API_URL="https://pulsechecks-api-prod-z5kj4psbnq-uc.a.run.app"
TF_ACC=1 go test ./internal/provider/... -v -timeout 120s
```

### Releasing

Tag a new version and push — GitHub Actions handles the rest:

```bash
git tag v0.2.0
git push origin v0.2.0
```

## License

[Mozilla Public License 2.0](LICENSE)
