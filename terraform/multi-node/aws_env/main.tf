# Setup provider
provider "aws" {
  region = "${var.aws_region}"
}

# Get VPC information from remote state
data "terraform_remote_state" "multinode_vpc" {
  backend = "s3"

  config {
    bucket = "${var.tf_bucket_name}"
    key    = "${var.vpc_state_file}"
    region = "${var.aws_region}"
  }
}

# Prepare separate subnet based on 
module "aws_subnet" {
  source = "git::ssh://git@github.com/intelsdi-x/terraform-kopernik//modules/aws_subnet"

  vpc_id      = "${data.terraform_remote_state.multinode_vpc.vpc_id}"
  vpc_cidr    = "${data.terraform_remote_state.multinode_vpc.vpc_cidr}"
  third_octet = "${var.agent_seed}"
}

# Open all ports inside private network.
resource "aws_security_group" "open_private_network" {
  vpc_id = "${data.terraform_remote_state.multinode_vpc.vpc_id}"
  description = "open network connectivity between nodes in private network."
  ingress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["${data.terraform_remote_state.multinode_vpc.vpc_cidr}"]
  }
}

# Create requested AWS instances.
resource "aws_instance" "cluster" {
  instance_type = "${var.aws_instance_type}"
  ami           = "${var.aws_ami}"
  key_name      = "${var.aws_keyname}"

  tags {
    Name       = "${element(var.hostnames, count.index)}"
    UserName   = "${element(var.user_name, count.index)}"
    PortNumber = "${element(var.port_number, count.index)}"
  }

  subnet_id              = "${module.aws_subnet.subnet_id}"
  vpc_security_group_ids = ["${module.aws_subnet.default_sg_id}", "${aws_security_group.open_private_network.id}"]
  count                  = "${var.count}"
}

# Helper variable, hold connection information for each VM
module "conn_vars" {
  source = "git::ssh://git@github.com/intelsdi-x/terraform-kopernik//modules/custom_var"

  input_map = "${map(
                 "remote_ip"    , "${aws_instance.cluster.*.public_ip}",
                 "remote_port"  , "${var.port_number}",
                 "pk_location"  , "${list("/dev/null")}",
                 "user"         , "${var.user_name}",
                 "type"         , "${var.conn_type}",
                 "use_ssh_agent", "${var.use_ssh_agent}"
  )}"
}

# Update hostnames & /etc/hosts on each VM
module "hosts_update" {
  source = "git::ssh://git@github.com/intelsdi-x/terraform-kopernik//modules/hosts_update"

  count           = "${var.count}"
  connection_info = "${module.conn_vars.return_map}"

  hostname_list = "${var.hostnames}"
  hosts_ip_list = ["${aws_instance.cluster.*.private_ip}"]
}

# Setup passwordless access between VMs
module "update_mesh" {
  source = "git::ssh://git@github.com/intelsdi-x/terraform-kopernik//modules/cluster_connection"

  count           = "${var.count}"
  connection_info = "${module.conn_vars.return_map}"
}

module "swan_deploy" {
  source = "../shared"

  count           = "${var.count}"
  hosts_ip_list   = ["${aws_instance.cluster.*.private_ip}"]
  connection_info = "${module.conn_vars.return_map}"
  repo_path       = "${var.repo_path}"
}
