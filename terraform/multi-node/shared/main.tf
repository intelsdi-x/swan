resource "null_resource" "deploying_swan" {
  count = "${var.count}"

  connection {
    type = "${element(var.connection_info["type"], count.index)}"
    user = "${element(var.connection_info["user"], count.index)}"

    host        = "${element(var.connection_info["remote_ip"], count.index)}"
    port        = "${element(var.connection_info["remote_port"], count.index)}"
    private_key = "${file(element(var.connection_info["pk_location"], count.index))}"
    agent       = "${element(var.connection_info["use_ssh_agent"], count.index)}"
  }

  # Prepare directories on remote environment.
  provisioner "remote-exec" {
    script = "${var.repo_path}/terraform/multi-node/shared/env_preparation.sh"
  }

  # Upload S3 auth file to remote environment.
  provisioner "local-exec" {
    command = "cat $HOME/swan_s3_creds/.s3cfg > .s3cfg" 
  }
  provisioner "file" {
    source = ".s3cfg"
    destination = "/home/${element(var.connection_info["user"], count.index)}/swan_s3_creds/.s3cfg"
  }

  # Upload vagrant deployment to remote environment.
  provisioner "file" {
    source = "${var.repo_path}/misc/dev/vagrant/singlenode/resources"
    destination = "/vagrant/"
  }

  # Upload artifacts uploader/downloader to remote environment.
  provisioner "file" {
    source = "${var.repo_path}/scripts/artifacts.sh"
    destination = "/vagrant/resources/scripts/artifacts.sh"
  }

  # Upload experiment configuartion to remote environment.
  provisioner "file" {
    source = "${var.repo_path}/terraform/multi-node/shared/experiment.sh"
    destination = "/home/${element(var.connection_info["user"], count.index)}/experiment.sh"
  }

  # Provisioning swan and its dependencies.
  provisioner "remote-exec" {
    script = "${var.repo_path}/terraform/multi-node/shared/swan_deploy.sh"
  }

}
