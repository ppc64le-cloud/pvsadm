# Overview
This guide talks about how to upload image file to Cloud object Storage instance using pvsadm.

# Prerequisite
- pvsadm tool
- Image file(RHEL/CENTOS/CoreOS)
- IBMCLOUD_API_KEY. [How to create api key](https://cloud.ibm.com/docs/account?topic=account-userapikey#create_user_key)

# Image upload command help:
```shell
$pvsadm image upload --help
Usage:
  pvsadm image upload [flags]

Flags:
      --resource-group string      Name of user resource group. (default "default")
      --cos-storageclass string    Cloud Object Storage Class type, available values are [standard, smart, cold, vault]. (default "standard")
  -n, --cos-instance-name string   Cloud Object Storage instance name.
  -b, --bucket string              Cloud Object Storage bucket name.
  -f, --file string                The PATH to the file to upload.
  -r, --bucket-region string       Cloud Object Storage bucket region. (default "us-south")
  -h, --help                       help for upload
```

# Image upload using pvsadm tool

Mandatory parameters for uploading the object to Cloud object storage using pvsadm tool are --bucket and --file

### case 1:
If user wants to upload the object to the bucket in the default region(us-south).
```shell
$pvsadm image upload --bucket bucket0711 -f rhcos-461.ova.gz
```

### case 2:
If bucket doesn't exists in default region(us-south), then user has to provide --bucket-region parameter to pvsadm tool. 
```shell
$pvsadm image upload --bucket bucket0711 -f rhcos-461.ova.gz --bucket-region <REGION>
```

### case 3:
If user knows Cloud Object Storage Instance where bucket is created, then use --cos-instance-name
```shell
$pvsadm image upload --bucket bucket0711 -f rhcos-461.ova.gz --cos-instance-name pvsadm-cos-instance --bucket-region <REGION>
```

### case 4:
If user is planning to create a new Cloud Object Storage Instance(S3 service)
```shell
$pvsadm image upload --bucket bucket1320 -f centos-8-latest.ova.gz --resource-group <ResourceGroup_Name> --bucket-region <REGION>
```

### case 5:
If user wants to upload the object to the bucket using access and secret key
```shell
$pvsadm image upload --bucket bucket1320 -f centos-8-latest.ova.gz  --bucket-region <REGION> --accesskey <ACCESSKEY> --secretkey <SECRETKEY>
```