terraform {
  required_providers {
    pulsechecks = {
      source  = "abdallah/pulsechecks"
      version = "~> 0.1"
    }
  }
}

provider "pulsechecks" {
  # api_token = "..." # or set PULSECHECKS_API_TOKEN env var
}

resource "pulsechecks_check" "daily_backup" {
  team_id        = "my-team-id"
  name           = "Daily Backup"
  check_type     = "heartbeat"
  period_seconds = 86400
  grace_seconds  = 3600
}

output "ping_url" {
  value = "https://pulsechecks-api-prod-z5kj4psbnq-uc.a.run.app/ping/${pulsechecks_check.daily_backup.token}"
}
