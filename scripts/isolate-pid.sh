#!/bin/bash
# See: https://intelsdi.atlassian.net/browse/SCE-259
# This script allows us to isolate PID namespace while running experiments
# Usage: isolate-pid.sh path/to/experiment/binary
# The script uses sudo so you should not be suprised

function display_error {
	echo -e "\e[41m$1\e[0m"
	exit 255
}

echo -e "\e[32mChecking unshare capabilities...\e[0m"
HAS_PID=`unshare -h | grep "\-\-pid"`
if [ "$HAS_PID" == "" ]; then
	echo -e "\e[32mRunning \e[7m$@\e[27m due to lack of PID namespace support in unshare\e[0m"	
	$@
	exit $?
fi

if [ "$#" -lt "1" ]; then
	display_error "Missing path to experiment binary. Example usage: \e[7m$0 path/to/experiment/binary\e[27ms"
        exit 251
fi

BINARY=$1
type -p $1 2>/dev/null 1>/dev/null
if [ "$?" -ne "0" ]; then
	display_error "$1 not found in PATH"
	exit 250
fi
BINARY=`type -p $1`

if [ ! -f "$BINARY" ]; then
	display_error "$BINARY does not exist or is not a file; it needs to be path to an existing file"
	exit 252
fi

if [ ! -x "$BINARY" ]; then
	display_error "$BINARY is not executable"
	exit 253
fi

read -ra ARGS <<< "$@"
ARGS[0]=$BINARY
EXEC="sudo -E unshare --pid --fork --mount-proc ${ARGS[*]}"

echo -e "\e[32mRunning \e[7m$EXEC\e[27m\e[0m"
$EXEC 

exit $?
