# Smoke fixture for --bind-service: the name below does not exist in any DNS.
# It resolves only because the module binds the caller's service under it, so
# a successful read proves the binding. Plain HTTP on purpose: a CDN front
# answers an unknown TLS SNI with silence, which would test the CDN, not this.
terraform {
  required_providers {
    http = {
      source  = "hashicorp/http"
      version = "~> 3.4"
    }
  }
}

variable "url" {
  type    = string
  default = "http://pinned.dagger-smoke.test"
}

data "http" "pinned" {
  url      = var.url
  insecure = true
}

output "status_code" {
  value = data.http.pinned.status_code
}
