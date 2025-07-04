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
  provisioner "local-exec" {
    command = "echo Hello, my name is ${var.name} and I like to eat ${var.food}."
  }
}

output "message" {
  value = "Terraform completed successfully for ${var.name}, who likes ${var.food}"
}
