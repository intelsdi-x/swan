output "mapped_IDs" {
  value = "${zipmap(vagrant_vbox.vbox.*.box_name, vagrant_vbox.vbox.*.id)}"
}

output "mapped_ips" {
  value = "${zipmap(vagrant_vbox.vbox.*.box_name, vagrant_vbox.vbox.*.remote_ip)}"
}

output "mapped_ports" {
  value = "${zipmap(vagrant_vbox.vbox.*.box_name, vagrant_vbox.vbox.*.remote_port)}"
}

output "mapped_users" {
  value = "${zipmap(vagrant_vbox.vbox.*.box_name, vagrant_vbox.vbox.*.remote_user)}"
}
