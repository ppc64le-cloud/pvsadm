# Overview
This guide talks about how to import image to PowerVs instance using pvsadm.

# Prerequisite
- pvsadm tool
- IBMCLOUD_API_KEY. [How to create api key](https://cloud.ibm.com/docs/account?topic=account-userapikey#create_user_key)
- S3 BucketName, Bucket Region, ObjectName
- PowerVS Instance Name/PowerVS Instance ID.

# Image import Command help
```shell
$pvsadm image import --help
Flags:
      --accesskey string                 Cloud Storage access key
  -b, --bucket string                    Cloud Storage bucket name
  -h, --help                             help for import
      --image-name string                Name to give imported image
  -i, --instance-id string               Instance ID of the PowerVS instance
  -n, --instance-name string             Instance name of the PowerVS
  -o, --object-name string               Cloud Storage image filename
  -r, --region string                    COS bucket location
  -p, --public-bucket                    Cloud Storage public bucket
      --secretkey string                 Cloud Storage secret key
      --service-credential-name string   Service Credential name to be auto generated (default "pvsadm-service-cred")
      --storagetype string               Storage type, accepted values are [tier1, tier3] (default "tier3")
  
```

# Importing Image to PowerVS instance from S3 Bucket using pvsadm

Set the API key variable
```shell
$export IBMCLOUD_API_KEY=<IBM_CLOUD_API_KEY>
```

### case 1:
Importing the image using auto-generated s3 credential
```shell
$pvsadm image import -n <POWERVS_INSTANCE_NAME> -b <BUCKETNAME> --object rhel-83-10032020.ova.gz --pvs-image-name test-image -r <REGION>
```

### case 2:
Importing the image using accesskey and secretkey
```shell
$pvsadm image import -n <POWERVS_INSTANCE_NAME> -b <BUCKETNAME> --accesskey <ACCESSKEY> --secretkey <SECRETKEY> --object rhel-83-10032020.ova.gz --pvs-image-name test-image -r <REGION>
```

### case 3:
If user wants to specify the PowerVS storage type for importing the image
```shell
$pvsadm image import -n <POWERVS_INSTANCE_NAME> -b <BUCKETNAME> --object rhel-83-10032020.ova.gz --pvs-image-name test-image -r <REGION> --storagetype <POWERVS_STORAGE_TYPE>
```

### case 4:
If user wants to specify type of OS 
```shell
$pvsadm image import -n <POWERVS_INSTANCE_NAME> -b <BUCKETNAME> --object rhel-83-10032020.ova.gz --pvs-image-name test-image -r <REGION>
```

### case 5:
Importing the image from public bucket 
```shell
$pvsadm image import -n <POWERVS_INSTANCE_NAME> -b <BUCKETNAME> --object rhel-83-10032020.ova.gz --pvs-image-name test-image -r <REGION> --public-bucket
```
