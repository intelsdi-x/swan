#!/bin/bash
# See: https://intelsdi.atlassian.net/browse/SCE-259
# This script allows us to isolate PID namespace while running experiments
# usage: sudo -E runs-experiment.sh path/to/experiment/binary
# -E option is quite important as we need to make sure that environment variables are passed to the experiment

function display_error {
	echo -e "\e[41m$1\e[0m"
	exit 255
}

if [ "$UID" -ne "0" ]; then
	display_error "You need to be root (hint: \e[7msudo -E $0 $1\e[27m)"
	exit 254
fi

if [ "$#" -ne "1" ]; then
	display_error "Missing path to experiment binary. Example usage: \e[7m$0 path/to/experiment/binary\e[27ms"
	exit 253
fi

if [ ! -f "$1" ]; then
	display_error "$1 does not exist or is not a file; it needs to be path to an existing file"
	exit 252
fi

if [ ! -x "$1" ]; then
	display_error "$1 is not executable"
	exit 253
fi

echo -e "\e[32mRunning $1 in separate PID namespace...\e[0m"
unshare --pid --fork --mount-proc $1
exit $?
