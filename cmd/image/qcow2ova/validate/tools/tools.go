package tools

import (
	"os/exec"

	"k8s.io/klog/v2"
)

var commands = map[string]string{
	"qemu-img": "yum install qemu-img -y",
	"growpart": "yum install cloud-utils-growpart -y",
}

type Rule struct {
	failedCommand string
}

func (p *Rule) String() string {
	return "tools"
}

func (p *Rule) Verify() error {
	for command := range commands {
		path, err := exec.LookPath(command)
		if err != nil {
			p.failedCommand = command
			return err
		}
		klog.Infof("%s found at %s\n", command, path)
	}
	return nil
}

func (p *Rule) Hint() string {
	if p.failedCommand != "" {
		return commands[p.failedCommand]
	}
	return ""
}
