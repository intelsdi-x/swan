# Remote multi-node experiment using Terraform and AWS\Vagrant

## Requirements

Shared requirements:
- [Terraform](https://www.terraform.io/)
- `.s3cfg` file placed in `$HOME/swan_s3_creds/`
  - File format:
  ```
  [default]
  access_key = <S3_ACCESS_KEY>
  secret_key = <S3_SECRET_KEY>
  ```
- Access to terraform modules provided by Kopernik team.
- Exported to S3 public key.
- (optional & by default) SSH agent to provide private key.
  - You can change this settings. Please look into [configuration](#configuration) section.


For local cluster(based on Vagrant):
- [Vagrant](https://www.vagrantup.com/)
- [VirtualBox](https://www.virtualbox.org/)
- [Terraform Vagrant Provider](https://github.com/intelsdi-x/terraform-provider-vagrant)

For remote cluster(created on AWS):
- Exported AWS credentials
- Exported `TF_VAR_agent_seed` with non-coliding third octed of subnet declaration.

## Quick Start

Depends on which cluster you want to build choose run this commands from:
- `aws_env` - for remote cluster
- `vagrant_env` - for local cluster

```
$ terraform get // this step will download required modules from GitHub
$ terraform plan // terraform integrity validation
$ terraform apply // building cluster
$ terraform outputs // display connection details for each node in cluster
$ ssh -p <PORT_NUMBER> <USER_NAME>@<IP_ADDRESS> // SSH connection will use your exported public key.
$ bash ./experiment.sh
```

## Configuration

Configuration is based on terraform variables (part of HCL language). Detailed informations about possibilities of HCL's variables are available [here](https://www.terraform.io/docs/configuration/variables.html)

There are three files with configuration for multi-node cluster based on terraform:
- `variables.tf` - shared configuration for remote and local cluster.
- `aws_env/env_variables.tf` - configuration for remote cluster based on AWS.
- `vagrant_env/env_variables.tf` - configuration for local cluster based on Vagrant.

### `variables.tf`

|variable|type|default|description|
|---|---|---|---|
|count|integer|3|Number of nodes in cluster|
|hostnames|list of strings|["controller", "node1", "node2", "node3", "node4", "node5"]|Hostnames for cluster nodes.<sup>\*</sup>|
|user_name|list of strings|["centos"]|Default user name which should be created inside VM.|
|port_number|list of strings|["22"]|Default connectivity port.<sup>\*\*</sup>|
|conn_type|list of strings|["ssh"]|Connection type which should be used. `ssh`/`winrm`|
|key_location|string|~/.ssh/id_rsa|Private key which should be used for provisioning connectivity.|
|repo_path|string|../../../|Relative path to repository root directory.<sup>\*\*\*</sup>|


<sup>\*</sup>Please provide at least <number_of_nodes_in_cluster> different hostnames.  
<sup>\*\*</sup>Terraform can use `ssh` or `winrm` for connectivity.  
<sup>\*\*\*</sup>This path should be changed for each terraform files movement. Relative path is counted from `aws_env`/`vagrant_env` directories.

### `aws_env/env_variables.tf`

|variable|type|default|description|
|---|---|---|---|
|aws_region|string|us-east-1|Existing region on AWS.|
|aws_ami|string|ami-29023a3e|AMI used for cluster deploting.|
|aws_instance_type|string|m4.4xlarge|Instance flavor.|
|aws_keyname|string|snapbot-private|Public key used for provisioning placed on AWS.<sup>\*</sup>|
|agent_seed|string|   |Non-coliding third octed of subnet declaration.|
|tf_bucket_name|string|terraform-statefiles|S3 bucket name which is storing .tfstate files(used for reviving broken cluster)|
|vpc_state_file|string|multinode_vpc/terraform.tfstate|Path on S3 bucket to VPC configuration.|
|use_ssh_agent|list of strings|["true"]|Use SSH agent for connectivity.|



<sup>\*</sup>This public key should be generated from private key pointed in [`variables.tf`](#variablestf)  

### `vagrant_env/env_variables.tf`

|variable|type|default|description|
|---|---|---|---|
|ips|list of strings|["10.11.12.13","10.11.12.14","10.11.12.15","10.11.12.16","10.11.12.17"]|Private IP addresses used inside cluster.<sup>\*</sup>|
|volumes|map[string]string|{"~/.glide"   = "/home/vagrant/.glide"}|Mapped volumes.|
|use_ssh_agent|list of strings|["true"]|Use SSH agent for connectivity.|

<sup>\*</sup>Please provide at least <number_of_nodes_in_cluster> different IP addresses.  
