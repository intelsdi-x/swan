#!/bin/bash

REPOSITORY_URL=github.com/intelsdi-x/swan
TARGET_PATH="/opt/gopath/src/github.com/intelsdi-x"

function printError() {
    echo -e "\033[0;31m### $1 ###\033[0m"
}

function printStep() {
    echo -e "\033[0;36m[$1]\033[0m"
}

function printInfo() {
    echo -e "\033[0;32m> $1\033[0m"
}

function verifyStatus() {
    if [[ !$? -eq 0 ]]; then
        printError "Step doesn't exit with exit code 0"
        cleanUp
        exit 1
    fi
}

function cleanUp() {
    rm -rf $(ls $TARGET_PATH | grep -v swan)
}

function setGitHubCredentials() {
    printStep "Set GitHub credentials"
    if [[ $GIT_TOKEN != "" ]]; then
        git config --global url."https://$GIT_TOKEN:x-oauth-basic@github.com/".insteadOf "https://github.com/"
        printInfo "Token for GitHub has been set"
    fi
}

function buildSnap() {
    printStep "Build snap"
    git clone https://github.com/intelsdi-x/snap
    cd snap
    make deps
    make all
    verifyStatus
    printInfo "Snap has been build"
    cd $TARGET_PATH
}

function cloneCode() {
    printStep "Clone source code from GitHub"

    if [[ $GIT_BRANCH == "" ]]; then
        GIT_BRANCH="master"
    fi

    printInfo "Selected branch: $GIT_BRANCH"
    git clone -b $GIT_BRANCH  https://$REPOSITORY_URL

    verifyStatus
    printInfo "Clone source code has been completed"
}

function prepareEnvironment() {
    printStep "Prepare environment"
    make deps
    verifyStatus
    printInfo "All dependencies have been downloaded"
}

function runTests() {
    printStep "Testing"
    DEFAULT_TARGETS=all
    if [[ "$@" != "" ]]; then
        DEFAULT_TARGETS="$@"
    fi

    printInfo "Selected targets: $DEFAULT_TARGETS"

    make $DEFAULT_TARGETS
    verifyStatus
}

function getCodeFromDir() {
    printStep "Binding source code from /swan to proper directory in \$GOPATH"
    mkdir swan
    mount -o bind /swan ./swan
    verifyStatus
    printInfo "Binding has been completed"
}

function getCode() {
    if [[ -d "/swan" ]]; then
        getCodeFromDir
    else
        cloneCode
    fi
}

function createDirs() {
    printStep "Creating required dirs"

    if [[ ! -d ${TARGET_PATH} ]]; then
        mkdir -p ${TARGET_PATH}
    fi
    cd ${TARGET_PATH}

    printInfo "Enter into: $(pwd)"
}

function main() {
    echo "createDirs\n"
    date
    createDirs
    echo "\n\nsetGitHubCredentials"
    date
    setGitHubCredentials
    echo "\n\nbuildSnap"
    date
    buildSnap
    date
    echo "\n\ngetCode"
    getCode
    echo "\n\ncd swan"
    date
    cd swan
    echo "\n\nprepareEnvironment"
    date
    prepareEnvironment
    date
    echo "\n\nrunTests"
    runTests "$@"
    echo "\n\ncleanUp"
    date
    cleanUp
    date
}

main "$@"
