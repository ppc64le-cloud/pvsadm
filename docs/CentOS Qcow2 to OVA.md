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

CentOS Qcow2 images are generally available in the location

### CentOS stream images:
CentOS 8: https://cloud.centos.org/centos/8-stream/ppc64le/images/

CentOS 9: https://cloud.centos.org/centos/9-stream/ppc64le/images/

### CentOS old image(deprecated):
https://cloud.centos.org/centos/8/ppc64le/images/

Download the required GenericCloud qcow2 image

## Step 2: Convert the qcow2 to ova

```shell
# Convert the CentOS8.3 qcow2 to ova format with installing all the prerequisites required for image to work in the IBM Power Systems Virtual Server
$ pvsadm image qcow2ova  --image-name centos-8-stream-09212022  --image-url ./CentOS-Stream-GenericCloud-8-20220913.0.ppc64le.qcow2 --image-dist centos
```
