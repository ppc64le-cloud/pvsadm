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

package dhcp

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"

	. "github.com/sayotte/iscdhcp"
)

var (
	gateway, networkID, file, nameservers, mtu string
	mutex                                      = &sync.Mutex{}
)

func ipv4MaskString(m []byte) string {
	if len(m) != 4 {
		panic("ipv4Mask: len must be 4 bytes")
	}

	return fmt.Sprintf("%d.%d.%d.%d", m[0], m[1], m[2], m[3])
}

const dhcpdTemplate = `
default-lease-time 600;
max-lease-time 7200;
ddns-update-style none;
authoritative;

%s
`

func doEvery(d time.Duration, f func()) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		f()
	}
}

type RoutersOption []net.IP

// IndentedString implements the method of the same name in the Statement interface
func (ro RoutersOption) IndentedString(prefix string) string {
	s := prefix + "option routers "
	for i := 0; i < len(ro)-1; i++ {
		s += ro[i].String() + ", "
	}
	return s + ro[len(ro)-1].String() + ";\n"
}

type InterfaceMTUOption string

// IndentedString implements the method of the same name in the Statement interface
func (io InterfaceMTUOption) IndentedString(prefix string) string {
	return prefix + fmt.Sprintf("option interface-mtu %s;\n", io)
}

func syncDHCPD() {
	c, err := client.NewClientWithEnv(pkg.Options.APIKey, pkg.Options.Environment, false)
	if err != nil {
		klog.Fatalf("failed to create a session with IBM cloud: %v", err)
	}

	pvmclient, err := client.NewPVMClientWithEnv(c, pkg.Options.WorkspaceID, "", "prod")
	if err != nil {
		klog.Fatalf("failed to create a PVM client: %v", err)
	}

	n, err := pvmclient.NetworkClient.Get(networkID)
	if err != nil {
		klog.Fatalf("failed to fetch network by ID: %v", err)
	}

	ipv4Addr, ipv4Net, err := net.ParseCIDR(*n.Cidr)
	if err != nil {
		klog.Fatalf("failed to ParseCIDR: %v", err)
	}
	subnetStmt := SubnetStatement{
		SubnetNumber: ipv4Addr,
		Netmask:      net.ParseIP(ipv4MaskString(ipv4Net.Mask)),
	}

	ports, err := pvmclient.NetworkClient.GetAllPorts(networkID)
	if err != nil {
		klog.Fatalf("failed to get the ports: %v", err)
	}

	var g = n.Gateway

	if gateway != "" {
		g = gateway
	}

	router := RoutersOption{net.ParseIP(g)}
	subnetStmt.Statements = append(subnetStmt.Statements, router)

	nameserver := DomainNameServersOption{}
	var nsList = n.DNSServers
	if nameservers != "" {
		nsList = strings.Split(nameservers, ",")
	}

	for _, ns := range nsList {
		nameserver = append(nameserver, net.ParseIP(ns))
	}
	subnetStmt.Statements = append(subnetStmt.Statements, nameserver)

	if mtu != "" {
		var mtu = InterfaceMTUOption(mtu)
		subnetStmt.Statements = append(subnetStmt.Statements, mtu)
	}

	for _, port := range ports.Ports {
		hs := HostStatement{
			Hostname: "host" + strings.Split(*port.IPAddress, ".")[3],
		}
		hs.Statements = append(hs.Statements, HardwareStatement{HardwareType: "ethernet", HardwareAddress: *port.MacAddress})
		hs.Statements = append(hs.Statements, FixedAddressStatement{net.ParseIP(*port.IPAddress)})
		subnetStmt.Statements = append(subnetStmt.Statements, hs)
	}

	content := fmt.Sprintf(dhcpdTemplate, subnetStmt.IndentedString(""))
	mutex.Lock()
	if err = os.WriteFile(file, []byte(content), 0644); err != nil {
		klog.Fatalf("cannot write to dhcp.conf file: %v", err)
	}
	mutex.Unlock()
}

var Cmd = &cobra.Command{
	Use:     "dhcp-sync",
	Short:   "dhcp-sync command",
	Long:    `dhcp-sync tool is a tool populating the dhcpd.conf file from the PowerVS network and restart the dhcpd service.`,
	GroupID: "dhcp",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if pkg.Options.WorkspaceID == "" {
			return fmt.Errorf("--workspace-id is required")
		}
		if pkg.Options.APIKey == "" {
			return fmt.Errorf("api-key can't be empty, pass the token via --api-key or set IBMCLOUD_API_KEY environment variable")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		go doEvery(2*time.Minute, syncDHCPD)

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			klog.Fatalf("cannot create a new fsNotify watcher: %v", err)
		}
		defer watcher.Close()

		done := make(chan bool)
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					klog.V(2).Infof("received an fsWatcher event: %v", event)
					if event.Op&fsnotify.Write == fsnotify.Write {
						klog.V(2).Infof("%s has been modified, proceeding to restart dhcpd service", event.Name)
						exitcode, out, err := utils.RunCMD("systemctl", "restart", "dhcpd")
						if exitcode != 0 {
							klog.Errorf("failed to restart the dhcpd service, exitcode: %d, stdout: %s, err: %s", exitcode, out, err)
						}
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					klog.Errorf("received an fsWatcher error: %v", err)
				}
			}
		}()

		err = watcher.Add(file)
		if err != nil {
			klog.Fatalf("cannot sync DHCP server: %v", err)
		}
		<-done
		return nil
	},
}

func init() {
	Cmd.Flags().StringVarP(&pkg.Options.WorkspaceID, "instance-id", "i", "", "Instance ID of the PowerVS instance")
	Cmd.Flags().MarkDeprecated("instance-id", "instance-id is deprecated, workspace-id should be used")
	Cmd.Flags().StringVarP(&pkg.Options.WorkspaceID, "workspace-id", "w", "", "Workspace ID of the PowerVS instance")
	Cmd.Flags().StringVar(&networkID, "network-id", "", "Network ID to be monitored")
	Cmd.Flags().StringVar(&file, "file", "/etc/dhcp/dhcpd.conf", "DHCP conf file")
	Cmd.Flags().StringVar(&gateway, "gateway", "", "Override the gateway value with")
	Cmd.Flags().StringVar(&nameservers, "nameservers", "", "Override the DNS nameservers")
	Cmd.Flags().StringVar(&mtu, "mtu", "", "Interface MTU value, e.g: 1450")
	_ = Cmd.MarkFlagRequired("network-id")
}
