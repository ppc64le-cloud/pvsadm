#### Overview
The `convert-upload-images-powervs` script helps to create RHEL/RHCOS images in PowerVS services. This script creates its own CentOS vm to do all the conversion tasks.

#### What this script does - High Level Steps
1. Spin up a CentOS instance in PowerVS using the default CentOS image [call Terraform code]
2. Convert given RHEL/RHCOS images from the spun up CentOS instance [pvsadm tool]
3. Upload and Import the RHEL/RHCOS OVA images from the CentOS instance [pvsadm tool]

#### Environments supported
1. Mac
2. Windows-10 WSL or Cygwin
3. Linux

#### Prerequisite
1. Access to any of the environments supported (mentioned in above section) with internet connectivity.
2. IBM Cloud API key to login to the service instance
3. Valid RedHat subscription (for RHEL image)

#### Image upload command help
```shell
Automation for creating RHCOS/RHEL images to PowerVS Services. This is a wrapper for pvsadm tool.

Usage:
  ./convert-upload-images-powervs [ --rhel-url <url> | --rhcos-url <url> ] --service-name  <service name> --region <bucket region> --cos-bucket <bucket name> --cos-resource-group <resource group>  --cos-instance-name <cos instance name>

Args:
      --service-name string         A list of PowerVS service instances with comma-separated(Mandatory)
      --region string               Object store bucket region(Mandatory)
      --cos-bucket string           Object store bucket name(Mandatory)
      --cos-resource-group string   COS resource group(Mandatory)
      --cos-instance-name string    COS instance name(Mandatory)
      --rhel-url url                url pointing to the RHEL qcow2 image(optional)
      --rhcos-url url               url pointing to the RHCOS qcow2 image(optional)
      --help                        help for upload
```
#### How to download and use script
1. Download the script
```shell
curl -O -fsSL https://raw.githubusercontent.com/ppc64le-cloud/pvsadm/main/samples/convert-upload-images-powervs/convert-upload-images-powervs
```
2. Give execute permission
```shell
 chmod +x ./convert-upload-images-powervs
```
3. Export Required KEYS and Credentials
```shell
export IBMCLOUD_API_KEY="<ibm cloud api key>"
export RHEL_SUBSCRIPTION_USERNAME="<redhat subscription username>" (only if you are creating RHEL image)
export RHEL_SUBSCRIPTION_PASSWORD="<redhat subscription password>" (only if you are creating RHEL image)
export RHEL_ROOT_PASSWORD="<RHEL root password>" (if the user doesnt set this variable, the script will generate a password. Only if you are creating RHEL image)
```
4a. Running the script directly
```shell
Example:
./convert-upload-images-powervs  --service-name  my-powervs-service  --region us-south --cos-bucket my-cos-bucket --cos-resource-group my-resource-group --cos-instance-name my-cos-instance --rhel-url  https://access.cdn.redhat.com/content/origin/files/sha256/4e/xxxxxx/rhel-8.3-ppc64le-kvm.qcow2\?user\=xxxxxxxxxxx\&_auth_\=xxxxxxx  --rhcos-url  https://mirror.openshift.com/pub/openshift-v4/ppc64le/dependencies/rhcos/4.7/4.7.0/rhcos-4.7.0-ppc64le-openstack.ppc64le.qcow2.gz
```
4b. Running the script from a container
```shell
Example:
docker run -it -e IBMCLOUD_API_KEY=$IBMCLOUD_API_KEY -e  RHEL_SUBSCRIPTION_USERNAME=$RHEL_SUBSCRIPTION_USERNAME -e RHEL_SUBSCRIPTION_PASSWORD=$RHEL_SUBSCRIPTION_PASSWORD -e RHEL_ROOT_PASSWORD=$RHEL_ROOT_PASSWORD quay.io/powercloud/image-upload:0.1 --service-name  my-powervs-service  --region us-south --cos-bucket my-cos-bucket --cos-resource-group my-resource-group --cos-instance-name my-cos-instance --rhel-url  https://access.cdn.redhat.com/content/origin/files/sha256/4e/xxxxxx/rhel-8.3-ppc64le-kvm.qcow2\?user\=xxxxxxxxxxx\&_auth_\=xxxxxxx  --rhcos-url  https://mirror.openshift.com/pub/openshift-v4/ppc64le/dependencies/rhcos/4.7/4.7.0/rhcos-4.7.0-ppc64le-openstack.ppc64le.qcow2.gz
```

!!! Note
User needs to provide any of RHEL/RHCOS urls.
If the RHEL url is pointing to access.cdn.redhat.com, please use a newly created one. User should have access to the support site to get the RHEL url.
Official RHCOS images can get from https://mirror.openshift.com/pub/openshift-v4/ppc64le/dependencies/rhcos/
The first service name in the list would be used for creating the CentOS VM with the default OVA image.
