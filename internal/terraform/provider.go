package terraform

// GenerateProvider generates the Dynatrace Terraform provider configuration.
func GenerateProvider() string {
	return `terraform {
  required_providers {
    dynatrace = {
      source  = "dynatrace-oss/dynatrace"
      version = "~> 1.0"
    }
  }
}

provider "dynatrace" {
  # Set via environment variables:
  # DYNATRACE_ENV_URL
  # DYNATRACE_API_TOKEN
}
`
}
