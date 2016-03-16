#!/bin/bash -e

# This is scripts are handled via Makefile - don't run them manually!
# See README.md for instructions.

SOURCEDIR=$1
EXPERIMENT=$2

BUILDDIR=$SOURCEDIR/build
ROOTFS=$BUILDDIR/rootfs
EXPERIMENT_MAIN=${SOURCEDIR}/experiments/${EXPERIMENT}/main.go

BUILDCMD="go build -a -ldflags -w "

echo
echo "****  Building SWAN experiment '${EXPERIMENT}'  ****"
echo

# Disable CGO for builds
export CGO_ENABLED=0

# Clean build bin dir
rm -rf $ROOTFS/*

# Make dir
mkdir -p $ROOTFS

# Build Swan experiment
echo "Source Dir = $SOURCEDIR"
echo "Building From: ${EXPERIMENT_MAIN}"
$BUILDCMD -o $ROOTFS/$EXPERIMENT ${EXPERIMENT_MAIN}
# TODO(bplotka): Make sure here that build pass succesfuly
echo "Build should be available from here: $ROOTFS/$EXPERIMENT"

chmod +x $ROOTFS/$EXPERIMENT