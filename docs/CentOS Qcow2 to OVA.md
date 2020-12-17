# Overview
This guide talks about how to convert the CentOS qcow2 to ova format

# Prerequisite
- The latest RHEL/CentOS ppc64le machine(virtual/baremetal) with enough diskspace with root access
- Packages:
    - qemu-img
    - cloud-utils-growpart
- pvsadm tool

# Steps
## Step 1: Download the CentOS qcow2 image
CentOS Qcow2 images are generally available in the location: https://cloud.centos.org/centos/8/ppc64le/images/, download the required GenericCloud qcow2 image

## Step 2: Convert the qcow2 to ova

```shell
# Convert the CentOS8.3 qcow2 to ova format with installing all the prerequisites required for image to work in the IBM Power Systems Virtual Server
$ pvsadm image qcow2ova  --image-name centos-83-12172020  --image-url ./CentOS-8-GenericCloud-8.3.2011-20201204.2.ppc64le.qcow2 --image-dist centos
```
