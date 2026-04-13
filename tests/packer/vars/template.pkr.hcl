variable "name" {
  type    = string
  default = "unset"
}

variable "environment" {
  type    = string
  default = "unset"
}

variable "owner" {
  type    = string
  default = "unset"
}

variable "api_token" {
  type      = string
  default   = "unset"
  sensitive = true
}

variable "db_password" {
  type      = string
  default   = "unset"
  sensitive = true
}

source "null" "vars" {
  communicator = "none"
}

build {
  sources = ["source.null.vars"]

  provisioner "shell-local" {
    inline = [
      "echo name=${var.name}",
      "echo environment=${var.environment}",
      "echo owner=${var.owner}",
      "echo api_token_len=${length(var.api_token)}",
      "echo db_password_len=${length(var.db_password)}",
    ]
  }
}
