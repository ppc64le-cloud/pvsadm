API_KEY=""  #Api key value
RHEL_URL="" #Rhel qcow2 url [Get KVM Guest image url from https://access.redhat.com/downloads/content/279/ver=/rhel---8/8.3/ppc64le/product-software]
RHCOS_URL="" #Rhcos qcow2 url [Eg: https://mirror.openshift.com/pub/openshift-v4/ppc64le/dependencies/rhcos/4.6/latest/rhcos-4.6.8-ppc64le-openstack.ppc64le.qcow2.gz]
CENTOS_URL="" #Centos qcow2 url [Eg: https://cloud.centos.org/centos/8/ppc64le/images/CentOS-8-GenericCloud-8.3.2011-20201204.2.ppc64le.qcow2]
RHEL_USERNAME="" #Rhel Username required for rhel image conversion
RHEL_PASSWORD="" #Rhel Password required for rhel image conversion
BUCKET_NAME="test-bucket" #Bucket name for uploading image. (Can be a new or an existing bucket)
BUCKET_REGION="us-south" #Region where the bucket is present
RESOURCE_GROUP="default" #Resource group required while uploading image
PVS_INSTANCE_NAME="" #PVS instance name where the image needs to be imported
