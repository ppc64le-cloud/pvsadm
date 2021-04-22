## Overview

This is a tool to help with managing of resources in [IBM Power Systems Virtual Server](https://www.ibm.com/cloud/power-virtual-server).

❗ There is no formal support for any problems with this repo. For issues please open a GitHub [issue](https://github.com/ppc64le-cloud/pvsadm/issues)

## Installation
1. Go to the [releases page](https://github.com/ppc64le-cloud/pvsadm/releases/)
2. Select the latest release and download the relevant binary under the Assets section.
3. Run the `pvsadm --help` command to check the available subcommands and the options.

## Image Management
Sub command under the pvsadm tool to perform image related tasks like image conversion, uploading and importing into the IBM Power Systems Virtual Server instances. For more information, refer to the `pvsadm image --help` command.

The typical image workflow comprises of the following steps:

1. Download the qcow2 image.
2. Convert the downloaded qcow2 image to ova using `pvsadm image qcow2ova` command.
3. Upload the ova image to IBM Cloud Object Store Bucket using `pvsadm image upload` command.
4. Import the ova image to IBM Power Systems Virtual Server instances using `pvsadm image import` command.

### 'How To' Guides
- How to convert CentOS qcow2 to ova image format - [guide](docs/CentOS%20Qcow2%20to%20OVA.md)
- How to convert RHEL qcow2 to ova image format - [guide](docs/RHEL%20Qcow2%20to%20OVA.md)
- How to convert RHCOS(Red Hat CoreOS) qcow2 to ova image format - [guide](docs/RHCOS%20Qcow2%20to%20OVA.md)
- Advanced scenarios for Qcow2 to ova image conversion - [guide](docs/Advanced%20Scenarios%20for%20Qcow2%20to%20OVA.md)
- How to import image to PowerVS instance from COS - [guide](docs/How%20to%20Import%20Image%20to%20PowerVS%20Instance.md)
- How to upload image to COS bucket using pvsadm - [guide](docs/How%20to%20Upload%20Image%20to%20COS.md)

### Samples
Please take a look at the [samples](samples/README.md)  folder for end-to-end examples.

### 
For bugs/enhancement requests etc. please open a GitHub [issue](https://github.com/ppc64le-cloud/pvsadm/issues)
