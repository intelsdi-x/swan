## Launching SPECjbb:

1. Install SPECjbb dependencies for Ubuntu or Centos with scripts (must be run as root)
`$SWAN_ROOT/workloads/web_serving/specjbb_deps_ubuntu.sh`,
`$SWAN_ROOT/workloads/web_serving/specjbb_deps_centos.sh`.
It will install java-1.8.0-openjdk on your machine.
1. Download and extract SPECjbb with a script (must be run as root)
`$SWAN_ROOT/scripts/get_specjbb.sh`.
It will download an iso file, mount it and copy its content into
`$SWAN_ROOT/workloads/web_serving/`.
1. SPECjbb can be run with a jar:
`$SWAN_ROOT/workloads/web_serving/specjbb/specjbb2015.jar`.
For details how to run it please look at
https://www.spec.org/jbb2015/docs/userguide.pdf

