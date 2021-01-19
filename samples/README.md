## Overview
This folder contains some end-to-end samples using the pvsadm tool

### 1. Image conversion and upload to pvs instance sample script
The image_create_and_upload_script helps to create ova image from qcow2 image, upload the images to a bucket (new/existing) and import them as boot image to pvs instance using the pvsadm tool
#### Prerequisite
1. Access to shell on IBM Power® logical partition (LPAR) running RHEL 8.x or CentOS 8.x with internet connectivity and minimum 250GB of free disk space
2. Need a valid RedHat subscription for the RHEL image conversion
#### How to download and use script
1. Download the script and config file
```
curl -O -fsSL https://raw.github.com/ppc64le-cloud/pvsadm/master/samples/image-create-and-upload-sample/image_create_and_upload_script.sh
curl -O -fsSL https://raw.github.com/ppc64le-cloud/pvsadm/master/samples/image-create-and-upload-sample/config.sh
```
2. Change values in the config.sh file as per description given in file
3. Run the script 
```
./image_create_and_upload_script.sh
```

### 2. Sample script for RHEL/RHCOS image conversion from qcow2 to ova and upload/import to PowerVS service instance
The image_convert_upload_e2e.sh script helps to create RHEL and RHCOS ova images from qcow2 images, upload the images to a bucket (new/existing) and import them as boot image to PowerVS instance using the pvsadm tool

#### What this script does - High Level Steps
1. Download CentOS OVA from IBM FTP server or Unicamp
2. Upload to IBM COS using pvsadm
3. Import to PowerVS instance using pvsadm
4. Spin up a CentOS instance using the OVA in PowerVS [call Terraform code]
5. SSH to the CentOS instance and run RHEL/RHCOS image conversion
6. Upload and Import the RHEL/RHCOS OVA image from the CentOS instance

#### Environments supported
1. Mac
2. Windows-10 WSL or Cygwin
3. Linux (IBM Cloud classic - CentOS or RHEL instance)

#### Prerequisite
1. Access to any of the environments supported (mentioned in above section) with internet connectivity and minimum 250GB of free disk space 
2. IBM Cloud API key to login to the service instance
3. Need a valid RedHat subscription for the RHEL image conversion

#### How to download and use script
1. Download the script
```
curl -O -fsSL https://raw.github.com/ppc64le-cloud/pvsadm/master/samples/image-create-and-upload-sample/image_convert_upload_e2e.sh
```
2. Give execute permission
```
 chmod 777 ./image_convert_upload_e2e.sh
```
3. Run the script
```
Example:
./image_convert_upload_e2e.sh --centos-url http://9.47.90.173:8080/centos/Centos-83.ova.gz --rhel-url https://access.cdn.redhat.com/content/origin/files/sha256/4e/4e5c7cccb5ea7622640c97725912222458af743cf994e48f10239734c6a2ee65/rhel-8.3-ppc64le-kvm.qcow2\?user\=e35f87692fc38ef0f0eb5404dc81a448\&_auth_\=1607860919_4f148af49719db3061bbfcdf2f02f845 --rhcos-url https://mirror.openshift.com/pub/openshift-v4/ppc64le/dependencies/rhcos/4.6/4.6.1/rhcos-4.6.1-ppc64le-openstack.ppc64le.qcow2.gz  --instance-name  ocp-validation-frankfurt-02 --region us-south  --cos-bucket bucket-validation-team --cos-resource-group ocp-validation-resource-group
```

Note:
User needs to provide the following parameters to run the script
```
1. centos-url : URL to download centos image from
2. rhel-url : URL to download RHEL qcow2 image (rhel-8.3-ppc64le-kvm.qcow2) from. Log in to the Redhat customer Portal ->  https://access.redhat.com/downloads/content/279/ver=/rhel---8/8.3/ppc64le/product-software and Choose 'Red Hat Enterprise Linux 8.3 Update KVM Guest Image' and copy the link.
3. rhcos-url : URL to download RHCOS qcow2 image from -> https://mirror.openshift.com/pub/openshift-v4/ppc64le/dependencies/rhcos/4.6/4.6.1/rhcos-4.6.1-ppc64le-openstack.ppc64le.qcow2.gz
4. instance-name : Name of the service instance that you have access to, in the PowerVS environment
5. region : Name of the region that you have access to, in the PowerVS environment
6. cos-bucket : Name of the Cloud Object Storage bucket that you have access to
7. cos-resource-group : Name of the Cloud Object Storage resource group that you have access to, in the PowerVS environment
```

While executing, the script prompts for the following input from the user. Make sure to provide proper values.
```
1. Select the private network to use
2. Enter a short name to identify the cluster (inf-node)
3. Enter RHEL subscription username for bastion nodes
4. Enter the password for above username
5. Do you want to use the default configuration for infra node? (memory=16g processors=1 count=1)
```
