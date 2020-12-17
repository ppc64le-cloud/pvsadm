# Scenarios
## Scenario 1: Create a small 20GB image for testing

```shell
$ pvsadm image qcow2ova  --image-name rhel-83-12182020  --image-url ./rhel-8.3-ppc64le-kvm.qcow2 --image-dist rhel --rhn-user jsmith --rhn-password re@llyASt0ngRHNPass0rd --image-size 20
```

## Scenario 2: Create an image with user selected root password for debug purpose

```shell
$ pvsadm image qcow2ova  --image-name rhel-83-12182020  --image-url ./rhel-8.3-ppc64le-kvm.qcow2 --image-dist rhel --rhn-user jsmith --rhn-password re@llyASt0ngRHNPass0rd --os-password someEasyPassword
```

## Scenario 3: Use the user defined directory for the temp directory(place used for image conversion)

```shell
$ pvsadm image qcow2ova  --image-name rhel-83-12182020  --image-url ./rhel-8.3-ppc64le-kvm.qcow2 --image-dist rhel --rhn-user jsmith --rhn-password re@llyASt0ngRHNPass0rd --temp-dir /home/jsmith
```
