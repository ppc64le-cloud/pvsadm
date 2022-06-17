#!/usr/bin/env bash
# Copyright 2021 IBM Corp
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

source config.sh

#-------------------------------------------------------------------------
# Standardize the OBJECT_NAME from the url in format<distro>-<version>-<mmddyyyy>.ova.gz
# and IMAGE_NAME as <distro>-<version>-<mmddyyyy>
#-------------------------------------------------------------------------
function standardize_object_name() {
  log "Formatting Object name $1"
  local object_original_url=$1
  local object_original_name=${object_original_url##*/}
  local object_extension=""
  local object_temp_name=""
  if echo $object_original_name | grep -q -i centos; then
    echo $object_original_name | grep 'ova.gz' >/dev/null
    [ $? -ne 0 ] && error "Unsupported file format"
    object_extension="ova.gz"
    object_temp_name=${object_original_name%.*.*}
    DISTRO="centos"
  elif echo $object_original_name | grep -q -i rhcos; then
    echo $object_original_name | grep 'qcow2.gz' >/dev/null
    [ $? -ne 0 ] && error "Unsupported file format"
    object_extension="ova.gz"
    object_temp_name=${object_original_name%.*.*}
    DISTRO="coreos"
  elif echo $object_original_name | grep -q -i rhel; then
    echo $object_original_name | grep 'qcow2' >/dev/null
    [ $? -ne 0 ] && error "Unsupported file format"
    object_extension="ova.gz"
    object_temp_name=${object_original_name%.*}
    DISTRO="rhel"
  fi
    IMAGE_NAME=$(echo $object_temp_name | sed -e 's/\.//g' -e 's/ppc64le//g' -e 's/openstack//g' -e 's/kvm//g'  -e 's/_/-/g' -e 's/---*//g' -e 's/\([0-9]\+\)-GenericCloud-//g'|tr '[:upper:]' '[:lower:]'|grep -o -E '^[a-z]*-[0-9]{2}')-$(date +%m%d%Y)
    OBJECT_NAME=$IMAGE_NAME.${object_extension}
}

#-------------------------------------------------------------------------
# Creates the ova image and import into pvs instance
#-------------------------------------------------------------------------
function create_and_upload_image() {
  local url=$1
  standardize_object_name $url
  echo "Convert $DISTRO image"
  pvsadm image qcow2ova  --image-name $IMAGE_NAME --image-url  $url --image-dist $DISTRO  --rhn-user $RHEL_USERNAME  --rhn-password $RHEL_PASSWORD
  echo "Uploading $DISTRO image"
  pvsadm image upload  --bucket $BUCKET_NAME --file $OBJECT_NAME --bucket-region $BUCKET_REGION --resource-group $RESOURCE_GROUP
  echo "Importing $DISTRO Image"
  pvsadm image import --pvs-instance-name $PVS_INSTANCE_NAME --bucket $BUCKET_NAME --object $OBJECT_NAME --pvs-image-name $IMAGE_NAME --bucket-region $BUCKET_REGION
}

dnf install -y  qemu-img cloud-utils-growpart
echo 'Initializing pvsadm tool !'
curl -sL https://raw.githubusercontent.com/ppc64le-cloud/pvsadm/main/get.sh | FORCE=1 bash
export IBMCLOUD_API_KEY=$API_KEY

if [ -n "$CENTOS_URL" ]; then
  create_and_upload_image $CENTOS_URL
fi
if [ -n "$RHCOS_URL" ]; then
  create_and_upload_image $RHCOS_URL
fi
if [ -n "$RHEL_URL" ]; then
  create_and_upload_image $RHEL_URL
fi
