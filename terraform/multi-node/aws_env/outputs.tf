output "mapped_IDs" {
  value = "${zipmap(aws_instance.cluster.*.tags.Name, aws_instance.cluster.*.id)}"
}

output "mapped_ips" {
  value = "${zipmap(aws_instance.cluster.*.tags.Name, aws_instance.cluster.*.public_ip)}"
}

output "mapped_ports" {
  value = "${zipmap(aws_instance.cluster.*.tags.Name, aws_instance.cluster.*.tags.PortNumber)}"
}

output "mapped_users" {
  value = "${zipmap(aws_instance.cluster.*.tags.Name, aws_instance.cluster.*.tags.UserName)}"
}
