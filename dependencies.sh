#!/bin/bash

: '

Copyright (C) 2020 IBM Corporation

Licensed under the Apache License, Version 2.0 (the “License”);
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an “AS IS” BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

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
    
    wget -q --spider http://ibm.com

    if [ $? -eq 0 ]; then
        echo "Internet connection: OK!"
    else
        echo "ERROR: please, check your internet connection."
        exit
    fi
}


function install_pvsadmin() {

    curl -s "https://api.github.com/repos/ppc64le-cloud/pvsadm/releases" >> /tmp/pvsadmin.json

    TAGS=($(cat /tmp/pvsadmin.json | jq -r '.[].tag_name'))

    PVSADM_VERSION=$TAGS

    rm -f /tmp/pvsadmin.json

    if command -v "pvsadm" &> /dev/null; then
        echo "pvsadm is already installed!"
        exit
    fi

    wget -q -O /usr/local/bin/pvsadm "https://github.com/ppc64le-cloud/pvsadm/releases/download/$PVSADM_VERSION/pvsadm-$DISTRO-$ARCH"
    chmod +x /usr/local/bin/pvsadm
    pvsadm version
}


function install_ibmcloud() {

    if command -v "ibmcloud" &> /dev/null; then
        echo "ibmcloud is already installed!"
        exit
    fi

    if [ $DISTRO == "linux" ]; then
        echo "Installing ibmcloud CLI on Linux..."
        curl -fsSL https://clis.cloud.ibm.com/install/linux | sh
        ibmcloud plugin install power-iaas
    fi

    if [ $DISTRO == "darwin" ]; then
        echo "Installing ibmcloud CLI on Mac..."
        curl -fsSL https://clis.cloud.ibm.com/install/osx | sh
	ibmcloud plugin install power-iaas
    fi    
}


function install_dependencies() {

    if [ $DISTRO == "linux" ]; then

        echo "Installing dependencies on Linux..."
        OS=$(cat /etc/os-release | grep -w "ID" | awk -F "=" '{print $2}' | tr -d "\"")

        if [ $OS == "ubuntu" ]; then
            apt-get install -y jq curl wget python3 python3-pip qemu-utils cloud-utils cloud-guest-utils
            pip3 install -U jinja2 boto3 PyYAML
        fi
        if [ $OS == "centos" ]; then
            dnf install -y jq curl wget python38 python38-pip qemu-img cloud-utils-growpart
            pip3 install -U jinja2 boto3 PyYAML
        fi
        if [ $OS == "rhel" ]; then
            RH_REGISTRATION=$(subscription-manager identity 2> /tmp/rhsubs.out; cat /tmp/rhsubs.out; rm -f /tmp/rhsubs.out)
            if [[ "$RH_REGISTRATION" == *"not yet registered"* ]]; then
                echo "Please, ensure your system is subscribed to RedHat."
                exit 1
            else
                dnf install -y jq curl wget python38 python38-pip qemu-img cloud-utils-growpart
                pip3 install -U jinja2 boto3 PyYAML
            fi
        fi
    fi

    # We do not install the .ova image creation requirements on MacOS.
    if [ $DISTRO == "darwin" ]; then
        echo "Installing ibmcloud CLI on Mac..."
        if ! command -v "brew" &> /dev/null; then
            /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install.sh)"
        fi
        brew install -U python@3.9
        brew install -U jq
        pip3 install -U jinja2 boto3 PyYAML
    fi
}

function run (){

    identify_os
    check_connectivity
    install_dependencies
    install_ibmcloud
    install_pvsadmin
}

run "$@"
