terraform {
  required_providers {
    null = {
      source = "hashicorp/null"
      version = "~> 3.0"
    }
  }
}

provider "null" {}

resource "null_resource" "example" {
  provisioner "local-exec" {
    command = "echo Hello from Terraform"
  }
}

output "message" {
  value = "Terraform completed successfully"
}