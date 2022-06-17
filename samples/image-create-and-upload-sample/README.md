#### Overview
The `image_create_and_upload_script` helps to create ova image from qcow2 image, upload the images to a bucket (new/existing) and import them as boot image to pvs instance using the pvsadm tool
#### Prerequisite
1. Access to shell on IBM Power® logical partition (LPAR) running RHEL 8.x or CentOS 8.x with internet connectivity and minimum 250GB of free disk space
2. Need a valid RedHat subscription for the RHEL image conversion
#### How to download and use script
1. Download the script and config file
```
curl -O -fsSL https://raw.github.com/ppc64le-cloud/pvsadm/main/samples/image-create-and-upload-sample/image_create_and_upload_script.sh
curl -O -fsSL https://raw.github.com/ppc64le-cloud/pvsadm/main/samples/image-create-and-upload-sample/config.sh
```
2. Change values in the config.sh file as per description given in file
3. Run the script
```
./image_create_and_upload_script.sh