variable "ips" {
  type = "list"

  default = ["10.11.12.13",
    "10.11.12.14",
    "10.11.12.15",
    "10.11.12.16",
    "10.11.12.17",
  ]
}

variable "volumes" {
  type = "map"

  default = {
    "/tmp/swan_glide" = "/home/vagrant/.glide"
  }
}

variable "use_ssh_agent" {
  type    = "list"
  default = ["true"]
}
