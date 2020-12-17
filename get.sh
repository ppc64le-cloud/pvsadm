#!/usr/bin/env bash

: '
    Copyright (C) 2020 IBM Corporation

    Rafael Sene <rpsene@br.ibm.com> - Initial implementation.

    This is a helper script that eases the dependency setup and 
    configuration for the pvsadm tool. 
'

# Trap ctrl-c and call ctrl_c()
trap ctrl_c INT

function ctrl_c() {
    echo "Bye!"
}

function identify_os() {

    local OS="$(uname -s)"

    case "${OS}" in
        Linux*)     DISTRO=linux;;
        Darwin*)    DISTRO=darwin;;
        Catalina*)  DISTRO=darwin;;
        *)          DISTRO="UNKNOWN:${OS}"
    esac

    ARCH=$(uname -m)

    if [ "$ARCH" == "amd64" ] || [ "$ARCH" == "x86_64" ]; then
        ARCH=amd64
    fi

    export $ARCH
    export $DISTRO
}

function check_connectivity() {
    
    curl --output /dev/null --silent --head --fail http://github.com
    if [ ! $? -eq 0 ]; then
        echo
        echo "ERROR: please, check your internet connection."
        exit
    fi
}

function install_pvsadmin() {

    if [[ "$1" == *"--force"* ]]; then
       if command -v "pvsadm" &> /dev/null; then
           rm -f /usr/local/bin/pvsadm
       fi
    fi
    
    LATEST_RELEASE=$(curl --silent "https://api.github.com/repos/ppc64le-cloud/pvsadm/releases/latest" | jq -r '.tag_name') 

    if command -v "pvsadm" &> /dev/null; then
        echo "pvsadm is already installed!"
	pvsadm version
        exit
    fi

    curl --progress-bar -LJ "https://github.com/ppc64le-cloud/pvsadm/releases/download/$LATEST_RELEASE/pvsadm-$DISTRO-$ARCH" --output /usr/local/bin/pvsadm

    chmod +x /usr/local/bin/pvsadm
    pvsadm version
}

function run (){

    if [ -z $1 ]; then
       echo
       echo "To replace an old version of pvsadm, run this script as ./get.sh --force"
       echo
    fi

    identify_os
    check_connectivity
    install_pvsadmin $1
}

run "$@"
