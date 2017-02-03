variable "aws_region" {
  default = "us-east-1"
}

# Centos (non-cached)
variable "aws_ami" {
  default = "ami-0125d417"
}

variable "aws_instance_type" {
  default = "m4.4xlarge"
}

variable "aws_keyname" {
  default = "snapbot-private"
}

variable agent_seed {}

variable "tf_bucket_name" {
  default = "terraform-statefiles"
}

variable "vpc_state_file" {
  default = "multinode_vpc/terraform.tfstate"
}

variable "use_ssh_agent" {
  type    = "list"
  default = ["true"]
}
