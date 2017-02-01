provider "vagrant" {
  path = "vagrant"
}

resource "vagrant_box" "box" {
  box = "squall0gd/centos"
}

resource "vagrant_vbox" "vbox" {
  depends_on = ["vagrant_box.box"]
  box        = "${vagrant_box.box.box}"

  cpus = 2
  mem  = 4096
  count      = "${var.count}"
  private_ip = "${element(var.ips, count.index)}"
  box_name   = "${element(var.hostnames, count.index)}"
  volumes    = "${var.volumes}"
}

# Helper variable, hold connection information for each VM
module "conn_vars" {
  source = "git::ssh://git@github.com/intelsdi-x/terraform-kopernik//modules/custom_var"

  input_map = "${map(
                 "remote_ip"    , "${vagrant_vbox.vbox.*.remote_ip}",
                 "remote_port"  , "${vagrant_vbox.vbox.*.remote_port}",
                 "pk_location"  , "${vagrant_vbox.vbox.*.remote_pk}",
                 "user"         , "${vagrant_vbox.vbox.*.remote_user}",
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
  hosts_ip_list = ["${vagrant_vbox.vbox.*.private_ip}"]
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
  connection_info = "${module.conn_vars.return_map}"
  hosts_ip_list = ["${vagrant_vbox.vbox.*.private_ip}"]
  repo_path       = "${var.repo_path}"
}
