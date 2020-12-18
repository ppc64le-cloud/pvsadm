# Overview
This guide talks about how to convert the RHEL qcow2 to ova format

# Prerequisite
- The latest RHEL/CentOS ppc64le machine(virtual/baremetal) with enough diskspace with root access
- Packages:
    - qemu-img
    - cloud-utils-growpart
- pvsadm tool
- RHN subscription

# Steps
## Step 1: Download the RHEL cloud image
- Goto Red Hat Customer Portal - https://access.redhat.com/downloads/content/279/ver=/rhel---8/8.3/ppc64le/product-software
- Find the latest `KVM Guest Image` and copy the link from the `Download Now` link

>Notes:
> - Web page got a session timeout, make sure your session is active while copying the link
> - Download link has a `&` char, make sure to add a proper escape character to download file properly

## Step 2: Convert the qcow2 to ova

```shell
# Convert the rhel8.3 Qcow2 image to ova format with installing all the prerequisites required for image to work in the IBM Power Systems Virtual Server
$ pvsadm image qcow2ova  --image-name rhel-83-12182020  --image-url ./rhel-8.3-ppc64le-kvm.qcow2 --image-dist rhel --rhn-user jsmith --rhn-password re@llyASt0ngRHNPass0rd
```
