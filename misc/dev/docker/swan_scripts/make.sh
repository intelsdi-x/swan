#!/bin/bash

function runAndPrepareMakeTarget() {
    printStep "Preparing environment for running 'make $SCENARIO'"
    buildSnap
    cd swan
    prepareEnvironment
    runScenarioMake
}

function runScenarioMake() {
    printStep "Running scenario: $SCENARIO"
    DEFAULT_TARGETS=all
    if [[ "$SCENARIO" != "" ]]; then
        DEFAULT_TARGETS="$@"
    fi

    printInfo "Selected targets: $DEFAULT_TARGETS"

    make $DEFAULT_TARGETS
    verifyStatus
}

function buildSnap() {
    printStep "Build snap"
    git clone https://github.com/intelsdi-x/snap
    cd snap
    make deps
    make all
    verifyStatus
    printInfo "Snap has been build"
    cd ..
}
