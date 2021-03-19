// Copyright 2021 IBM Corp
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prep

import (
	"bytes"
	"fmt"
	"text/template"
)

// TODO: add a logic to make the package versions as an argument
var SetupTemplate = `#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail

mv /etc/resolv.conf /etc/resolv.conf.orig | true
echo "nameserver 9.9.9.9" | tee /etc/resolv.conf
{{if eq .Dist "rhel"}}
subscription-manager register --force --auto-attach --username={{ .RHNUser }} --password={{ .RHNPassword }}
{{end}}
yum update -y && yum install -y yum-utils
yum install -y cloud-init
ln -s /usr/lib/systemd/system/cloud-init-local.service /etc/systemd/system/multi-user.target.wants/cloud-init-local.service
ln -s /usr/lib/systemd/system/cloud-init.service /etc/systemd/system/multi-user.target.wants/cloud-init.service
ln -s /usr/lib/systemd/system/cloud-config.service /etc/systemd/system/multi-user.target.wants/cloud-config.service
ln -s /usr/lib/systemd/system/cloud-final.service /etc/systemd/system/multi-user.target.wants/cloud-final.service
rm -rf /etc/systemd/system/multi-user.target.wants/firewalld.service
rpm -vih --nodeps http://public.dhe.ibm.com/software/server/POWER/Linux/yum/download/ibm-power-repo-latest.noarch.rpm
sed -i 's/^more \/opt\/ibm\/lop\/notice/#more \/opt\/ibm\/lop\/notice/g' /opt/ibm/lop/configure
echo 'y' | /opt/ibm/lop/configure
# Disable the AT repository due to slowness in nature
yum-config-manager --disable Advance_Toolchain
yum install  powerpc-utils librtas DynamicRM  devices.chrp.base.ServiceRM rsct.opt.storagerm rsct.core rsct.basic rsct.core src -y
yum install -y device-mapper-multipath
cat <<EOF > /etc/multipath.conf
defaults {
    user_friendly_names yes
    verbosity 6
    polling_interval 10
    max_polling_interval 50
    reassign_maps yes
    failback immediate
    rr_min_io 2000
    no_path_retry 10
    checker_timeout 30
    find_multipaths smart
}
EOF
sed -i 's/GRUB_TIMEOUT=.*$/GRUB_TIMEOUT=60/g' /etc/default/grub
sed -i 's/GRUB_CMDLINE_LINUX=.*$/GRUB_CMDLINE_LINUX="console=tty0 console=hvc0,115200n8  biosdevname=0  crashkernel=auto rd.shell rd.debug rd.driver.pre=dm_multipath log_buf_len=1M "/g' /etc/default/grub
echo 'force_drivers+=" dm-multipath "' >/etc/dracut.conf.d/10-mp.conf
dracut --regenerate-all --force
for kernel in $(rpm -q kernel | sort -V | sed 's/kernel-//')
do
	echo "Generating initramfs for kernel version: ${kernel}"
	dracut --kver ${kernel} --force --add multipath --include /etc/multipath /etc/multipath --include /etc/multipath.conf /etc/multipath.conf
done
grub2-mkconfig -o /boot/grub2/grub.cfg
rm -rf /etc/sysconfig/network-scripts/ifcfg-eth0
echo {{ .RootPasswd }} | passwd root --stdin
{{if eq .Dist "rhel"}}
subscription-manager unregister
subscription-manager clean
{{end}}

# Remove the ibm repositories used for the rsct installation
rpm -e ibm-power-repo-*.noarch

mv /etc/resolv.conf.orig /etc/resolv.conf | true
touch /.autorelabel
`

var cloudConfig = `# latest file from cloud-init-19.4-11.el8_3.2.noarch
users:
 - default

## Change 1: Enabling the root login
disable_root: 0
ssh_pwauth:   0

mount_default_fields: [~, ~, 'auto', 'defaults,nofail,x-systemd.requires=cloud-init.service', '0', '2']
resize_rootfs_tmp: /dev
ssh_deletekeys:   1
ssh_genkeytypes:  ~
syslog_fix_perms: ~
disable_vmware_customization: false

cloud_init_modules:
 - disk_setup
 - migrator
 - bootcmd
 - write-files
 - growpart
 - resizefs
 - set_hostname
 - update_hostname
 - update_etc_hosts
 - rsyslog
 - users-groups
 - ssh

cloud_config_modules:
 - mounts
 - locale
 - set-passwords
 - rh_subscription
 - yum-add-repo
 - package-update-upgrade-install
 - timezone
 - puppet
 - chef
 - salt-minion
 - mcollective
 - disable-ec2-metadata
 - runcmd

cloud_final_modules:
 - rightscale_userdata
 - scripts-per-once
 - scripts-per-boot
 - scripts-per-instance
 - scripts-user
 - ssh-authkey-fingerprints
 - keys-to-console
 - phone-home
 - final-message
 - power-state-change
 - reset_rmc
### ^^^ Change 2: Recommendation from PowerVC

system_info:
  default_user:
    name: cloud-user
    lock_passwd: true
    gecos: Cloud User
    groups: [adm, systemd-journal]
    sudo: ["ALL=(ALL) NOPASSWD:ALL"]
    shell: /bin/bash
  distro: rhel
  paths:
    cloud_dir: /var/lib/cloud
    templates_dir: /etc/cloud/templates
  ssh_svcname: sshd

###############################################
### Change 3: Recommendation from PowerVC######
datasource_list: [ ConfigDrive, NoCloud, None ]
datasource:
  ConfigDrive:
    dsmode: local
###############################################

######################################################################################################################
### Change 4: Disable the network config in cloud-init post deployment, known issue when deployed with multipath disks
write_files:
- path: /usr/local/bin/disable_cloud_init_nw.sh
  permissions: 0755
  owner: root
  content: |
    #!/usr/bin/env bash
    set -e
    cat <<EOF > /etc/cloud/cloud.cfg.d/01_disable_cloud_nw.cfg
    #cloud-config
    network:
      config: disabled
    EOF

runcmd:
    - bash /usr/local/bin/disable_cloud_init_nw.sh

######################################################################################################################

# vim:syntax=yaml
`

var dsIdentify = `policy: search,found=all,maybe=all,notfound=disabled
`

type Setup struct {
	Dist, RHNUser, RHNPassword, RootPasswd string
}

func Render(dist, rhnuser, rhnpasswd, rootpasswd string) (string, error) {
	s := Setup{
		dist, rhnuser, rhnpasswd, rootpasswd,
	}
	var wr bytes.Buffer
	t := template.Must(template.New("setup").Parse(SetupTemplate))
	err := t.Execute(&wr, s)
	if err != nil {
		return "", fmt.Errorf("error while rendoring the script template: %v", err)
	}
	return wr.String(), nil
}
