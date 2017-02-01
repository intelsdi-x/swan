variable connection_info {
  description = <<EOF
  map[string][list] of variables needed to establish connection to remote machines.
	Following keys must exist:
	 - remote_ip       - list of IPs to connect to
	 - remote_port     - list of ports to be used during connection
	 - pk_location     - list of private keys that should be used for authentication
	 - user            - list of users that should be used during login
	 - type            - connection type for each machine `ssh` or `winrm`
	 - use_ssh_agent   - use ssh agent for authentication `true` or `false`
EOF

  type = "map"
}

variable "count" {
  description = "number of machines to connect to"
}

variable hosts_ip_list {
  description = "list of private IPs"
  type        = "list"
}

variable "repo_path" {
  description = "Path to swan repository"
}
