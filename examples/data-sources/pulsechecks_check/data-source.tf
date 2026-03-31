terraform {
  required_providers {
    pulsechecks = {
      source  = "abdallah/pulsechecks"
      version = "~> 0.1"
    }
  }
}

provider "pulsechecks" {}

data "pulsechecks_check" "existing" {
  team_id  = "my-team-id"
  check_id = "abc123"
}

output "check_name" {
  value = data.pulsechecks_check.existing.name
}
