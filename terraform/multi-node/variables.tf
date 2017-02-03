variable "count" {
  default = 2
}

variable "hostnames" {
  type    = "list"
  default = ["controller", "node1", "node2", "node3", "node4", "node5"]
}

variable "user_name" {
  type    = "list"
  default = ["centos"]
}

variable "port_number" {
  type    = "list"
  default = ["22"]
}

variable "conn_type" {
  type    = "list"
  default = ["ssh"]
}

variable "key_location" {
  type    = "string"
  default = "~/.ssh/id_rsa"
}

variable "repo_path" {
  type = "string"
  # This is relative path to swan repository from current directory.
  # We cannot use path with GOPATH due to Kopernik workflow(which clone repo
  # to workspace directory).
  default = "../../../"
}
