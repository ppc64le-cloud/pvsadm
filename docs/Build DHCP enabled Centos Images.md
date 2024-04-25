# Overview

This guide talks about how to build centos image to support DHCP IP and import it to PowerVS workspace.

1. Build the image template
    ```
   pvsadm image qcow2ova --prep-template-default > image-prep.template
   ```
   
2. Add the below lines to end of `image-prep.template` file to customize network.
   ```shell
    mkdir -p /etc/cloud/cloud.cfg.d
    cat <<EOF >> /etc/cloud/cloud.cfg.d/99-custom-networking.cfg
    network: {config: disabled}
    EOF
   ```

3. Build OVA image
    ```
   pvsadm image qcow2ova --image-name <name> --image-dist centos --image-url https://cloud.centos.org/centos/8-stream/ppc64le/images/CentOS-Stream-GenericCloud-8-latest.ppc64le.qcow2 --prep-template image-prep.template
    ```
   Note: Qcow2 CentOS images for ppc64le can be found [here](https://cloud.centos.org/centos/8-stream/ppc64le/images/)


4. Upload image to COS bucket
    ```
   pvsadm image upload -b <bucket-name> -f <file-name> -r <region> --accesskey <access-key-value> --secretkey <secret-key-value>
   ```
5. Import  OVA image to a PowerVS Workspace
    ```
   pvsadm image import -n <service-instance-name> -b <bucket-name> -o <file-name> -r <region> --accesskey <access-key-value> --secretkey <secret-key-value> --pvs-image-name <final-image-name>
   ```