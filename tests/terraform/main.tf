terraform {
  required_providers {
    null = {
      source  = "hashicorp/null"
      version = "~> 3.0"
    }
  }
}

provider "null" {}

variable "name" {
  description = "Name to greet"
  type        = string
  default     = "luke"
}

variable "food" {
  description = "Favorite food"
  type        = string
  default     = "schnitzel"
}

resource "null_resource" "example" {
  # A changed name REPLACES the resource: the smoke test uses that to prove
  # refuse-destroy stops a plan that deletes.
  triggers = {
    name = var.name
  }

  provisioner "local-exec" {
    command = "echo Hello, my name is ${var.name} and I like to eat ${var.food}."
  }
}

# Only here so the smoke test can show --targets leaves it out.
resource "null_resource" "untargeted" {}

output "message" {
  value = "Terraform completed successfully for ${var.name}, who likes ${var.food}"
}
