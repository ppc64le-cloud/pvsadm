# pvsadm

## Overview
Power Systems Virtual Server projects deliver flexible compute capacity for Power Systems workloads.
Integrated with the IBM Cloud platform for on-demand provisioning.

This is a tool built for the Power Systems Virtual Server helps managing and maintaining the resources easily

```shell script
./pvsadm --help
Power Systems Virtual Server projects deliver flexible compute capacity for Power Systems workloads.
Integrated with the IBM Cloud platform for on-demand provisioning.

This is a tool built for the Power Systems Virtual Server helps managing and maintaining the resources easily

Usage:
  pvsadm [command]

Available Commands:
  help        Help about any command
  purge       Purge the powervs resources

Flags:
  -k, --api-key string   IBMCLOUD API Key(env name: IBMCLOUD_API_KEY)
      --debug            Enable PowerVS debug option
  -h, --help             help for pvsadm

Use "pvsadm [command] --help" for more information about a command.
```

## Purge PowerVS Resources:

```shell script
$ ./pvsadm purge --help
Purge the powervs resources

Usage:
  pvsadm purge [command]

Available Commands:
  images      Purge the powervs images
  networks    Purge the powervs networks
  vms         Purge the powervs vms
  volumes     Purge the powervs volumes

Flags:
      --before duration        Remove resources before mentioned duration(format: 99h99m00s), mutually exclusive with --since
      --dry-run                dry run the action and don't delete the actual resources
  -h, --help                   help for purge
      --ignore-errors          Ignore any errors during the operations
  -i, --instance-id string     Instance ID of the PowerVS instance
  -n, --instance-name string   Instance name of the PowerVS
      --no-prompt              Show prompt before doing any destructive operations
      --since duration         Remove resources since mentioned duration(format: 99h99m00s), mutually exclusive with --before

Global Flags:
  -k, --api-key string   IBMCLOUD API Key(env name: IBMCLOUD_API_KEY)
      --debug            Enable PowerVS debug option

Use "pvsadm purge [command] --help" for more information about a command.
```
