#!/bin/bash

set -e -o pipefail

usage="$(basename "$0") [-h] [-p] -- program to run sensitivity experiment

where:
    -h Print usage
    -p Path to experiment binary"

while getopts ':hp:' opt; do
  case "$opt" in
    h)
      echo "$usage"
      exit
      ;;
    p)
      path=$OPTARG
      ;;
    :)
      echo "Option -$OPTARG requires an argument." >&2
      exit 1
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      exit 1
      ;;
    esac
  done

ID=$($GOPATH/src/github.com/intelsdi-x/swan/$path)
if [ -z ${ID+x} ];
  then echo "could not retrieve experiment ID";
  else $GOPATH/src/github.com/intelsdi-x/swan/build/viewer/sensitivity_viewer sensitivity $ID;
fi
