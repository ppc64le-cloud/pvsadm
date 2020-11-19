package prep

import (
	"bytes"
	"fmt"
	"text/template"
)

// TODO: add a logic to make the package versions as an argument
var setup = `#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail
set -o xtrace
mv /etc/resolv.conf /etc/resolv.conf.orig | true
echo "nameserver 9.9.9.9" | tee /etc/resolv.conf
{{if eq .Dist "rhel"}}
subscription-manager register --force --auto-attach --username={{ .RHNUser }} --password={{ .RHNPassword }}
{{end}}
yum update -y
# yum install http://public.dhe.ibm.com/systems/virtualization/powervc/rhel8_cloud_init/cloud-init-19.1-8.ibm.el8.noarch.rpm -y
yum install http://people.redhat.com/~eterrell/cloud-init/cloud-init-19.4-11.el8_3.1.noarch.rpm -y
ln -s /usr/lib/systemd/system/cloud-init-local.service /etc/systemd/system/multi-user.target.wants/cloud-init-local.service
ln -s /usr/lib/systemd/system/cloud-init.service /etc/systemd/system/multi-user.target.wants/cloud-init.service
ln -s /usr/lib/systemd/system/cloud-config.service /etc/systemd/system/multi-user.target.wants/cloud-config.service
ln -s /usr/lib/systemd/system/cloud-final.service /etc/systemd/system/multi-user.target.wants/cloud-final.service
rm -rf /etc/systemd/system/multi-user.target.wants/firewalld.service
rpm -vih --nodeps http://public.dhe.ibm.com/software/server/POWER/Linux/yum/download/ibm-power-repo-latest.noarch.rpm
sed -i 's/^more \/opt\/ibm\/lop\/notice/#more \/opt\/ibm\/lop\/notice/g' /opt/ibm/lop/configure
echo 'y' | /opt/ibm/lop/configure
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
mv /etc/resolv.conf.orig /etc/resolv.conf | true
touch /.autorelabel`

var cloudConfig = `# The top level settings are used as module
# and system configuration.
# A set of users which may be applied and/or used by various modules
# when a 'default' entry is found it will reference the 'default_user'
# from the distro configuration specified below
users:
   - default
# If this is set, 'root' will not be able to ssh in and they
# will get a message to login instead as the default $user
disable_root: false
mount_default_fields: [~, ~, 'auto', 'defaults,nofail', '0', '2']
resize_rootfs_tmp: /dev
ssh_pwauth:   0
# This will cause the set+update hostname module to not operate (if true)
preserve_hostname: false
# Example datasource config
# datasource:
#    Ec2:
#      metadata_urls: [ 'blah.com' ]
#      timeout: 5 # (defaults to 50 seconds)
#      max_wait: 10 # (defaults to 120 seconds)
datasource_list: [ ConfigDrive, NoCloud, None ]
datasource:
  ConfigDrive:
    dsmode: local
# The modules that run in the 'init' stage
cloud_init_modules:
 - migrator
 - seed_random
 - bootcmd
 - write-files
 - growpart
 - resizefs
 - disk_setup
 - mounts
 - set_hostname
 - update_hostname
 - update_etc_hosts
 - ca-certs
 - rsyslog
 - users-groups
 - ssh
# The modules that run in the 'config' stage
cloud_config_modules:
 - ssh-import-id
 - locale
 - set-passwords
 - spacewalk
 - yum-add-repo
 - ntp
 - timezone
 - disable-ec2-metadata
 - runcmd
# The modules that run in the 'final' stage
cloud_final_modules:
 - package-update-upgrade-install
 - puppet
 - chef
 - mcollective
 - salt-minion
 - rightscale_userdata
 - scripts-vendor
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
# System and/or distro specific settings
# (not accessible to handlers/transforms)
system_info:
   # This will affect which distro class gets used
   distro: rhel
   # Default user name + that default users groups (if added/used)
   default_user:
     name: rhel
     lock_passwd: True
     gecos: rhel Cloud User
     groups: [wheel, adm, systemd-journal]
     sudo: ["ALL=(ALL) NOPASSWD:ALL"]
     shell: /bin/bash
   # Other config here will be given to the distro class and/or path classes
   paths:
      cloud_dir: /var/lib/cloud/
      templates_dir: /etc/cloud/templates/
   ssh_svcname: sshd

bootcmd:
    - 'echo "IPV6_AUTOCONF=no" >> /etc/sysconfig/network-scripts/ifcfg-$(ls  /sys/class/net -1| grep env.|sort -n -r|head -1)'`

var dsIdentify = `policy: search,found=all,maybe=all,notfound=disabled`

type Setup struct {
	Dist, RHNUser, RHNPassword, RootPasswd string
}

func Render(dist, rhnuser, rhnpasswd, rootpasswd string) (string, error) {
	s := Setup{
		dist, rhnuser, rhnpasswd, rootpasswd,
	}
	var wr bytes.Buffer
	t := template.Must(template.New("setup").Parse(setup))
	err := t.Execute(&wr, s)
	if err != nil {
		return "", fmt.Errorf("error while rendoring the script template: %v", err)
	}
	return wr.String(), nil
}
