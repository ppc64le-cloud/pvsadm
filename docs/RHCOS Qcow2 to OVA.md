# Overview
This guide talks about how to convert the Red Hat CoreOS qcow2 to ova format

# Prerequisite
- The latest RHEL/CentOS ppc64le machine(virtual/baremetal) with enough diskspace with root access
- Packages:
    - qemu-img
    - cloud-utils-growpart
- pvsadm tool

# Steps
## Step 1: RHCOS image
Red Hat CoreOS images can be found at the [link](https://mirror.openshift.com/pub/openshift-v4/ppc64le/dependencies/rhcos/)

>Note: Use the image ending with openstack.ppc64le.qcow2.gz
## Step 2: Convert the RHCOS qcow2 to ova

```shell
# Convert the rhcos 4.6 qcow2 to ova format
$ pvsadm image qcow2ova  --image-name rhcos-46  --image-url https://mirror.openshift.com/pub/openshift-v4/ppc64le/dependencies/rhcos/4.6/4.6.1/rhcos-4.6.1-ppc64le-openstack.ppc64le.qcow2.gz --image-dist coreos
```
