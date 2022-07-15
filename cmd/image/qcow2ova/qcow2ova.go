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

package qcow2ova

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/manifoldco/promptui"
	"github.com/ppc64le-cloud/pvsadm/cmd/image/qcow2ova/ova"
	"github.com/ppc64le-cloud/pvsadm/cmd/image/qcow2ova/prep"
	"github.com/ppc64le-cloud/pvsadm/cmd/image/qcow2ova/validate"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var Cmd = &cobra.Command{
	Use:   "qcow2ova",
	Short: "Convert the qcow2 image to ova format",
	Long: `Convert the qcow2 image to ova format

Examples:

  # Downloads the coreos image from remote site and converts into ova type with name rhcos-461.ova.gz
  pvsadm image qcow2ova --image-name rhcos-461 --image-dist coreos --image-url https://mirror.openshift.com/pub/openshift-v4/ppc64le/dependencies/rhcos/4.6/4.6.1/rhcos-4.6.1-ppc64le-openstack.ppc64le.qcow2.gz

  # Converts the CentOS image from the local filesystem with size 50GB
  pvsadm image qcow2ova --image-name centos-82 --image-dist centos --image-size 50 --image-url /root/CentOS-8-GenericCloud-8.2.2004-20200611.2.ppc64le.qcow2

  # Converts the RHEL image from local filesystem
  pvsadm image qcow2ova --image-name rhel-82-29oct --image-dist rhel --rhn-user joesmith@example.com --rhn-password someValidPassword --image-url ./rhel-8.2-update-2-ppc64le-kvm.qcow2

  # Converts the CentOS image from the local filesystem with OS password set
  pvsadm image qcow2ova --image-name centos-82 --image-dist centos --os-password s0meC0mplexPassword --image-url /root/CentOS-8-GenericCloud-8.2.2004-20200611.2.ppc64le.qcow2

  # Converts the CentOS image from the local filesystem without OS password
  pvsadm image qcow2ova --image-name centos-82 --image-dist centos  --image-url /root/CentOS-8-GenericCloud-8.2.2004-20200611.2.ppc64le.qcow2 --skip-os-password

  # Customize the image preparation script for RHEL/CentOS distro, e.g: add additional yum repository or packages, change name servers etc. 
  # Step 1 - Dump the default image preparation template
  pvsadm image qcow2ova --prep-template-default > image-prep.template
  # Step 2 - Make the necessary changes to the above generated template file(bash shell script) - image-prep.template
  # Step 3 - Run the qcow2ova with the modified image preparation template
  pvsadm image qcow2ova --image-name centos-82 --image-dist centos --image-url /root/CentOS-8-GenericCloud-8.2.2004-20200611.2.ppc64le.qcow2 --prep-template image-prep.template
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		opt := pkg.ImageCMDOptions

		if opt.PrepTemplateDefault {
			fmt.Println(prep.SetupTemplate)
			os.Exit(0)
		}

		// Override the prep.SetupTemplate if --prep-template supplied
		if opt.PrepTemplate != "" {
			if strings.ToLower(opt.ImageDist) == "coreos" {
				return fmt.Errorf("--prep-template option is not supported for coreos distro")
			} else {
				klog.Info("Overriding with the user defined image preparation template.")
				content, err := ioutil.ReadFile(opt.PrepTemplate)
				if err != nil {
					return err
				}
				prep.SetupTemplate = string(content)
			}
		}

		if !utils.Contains([]string{"rhel", "centos", "coreos"}, strings.ToLower(opt.ImageDist)) {
			klog.Errorln("--image-dist is a mandatory flag and one of these [rhel, centos, coreos]")
			os.Exit(1)
		}

		//Read the RHNUser and RHNPassword if empty
		if opt.ImageDist == "rhel" && (opt.RHNUser == "" || opt.RHNPassword == "") {
			var err error
			klog.Warning("rhn-user and rhn-password options are mandatory when image-dist is rhel, please enter the details")

			//Validates and make sure input is not an empty string
			validate := func(input string) error {
				if len(strings.TrimSpace(input)) == 0 {
					return fmt.Errorf("input can't be empty string")
				}
				return nil
			}
			if opt.RHNUser == "" {
				prompt := promptui.Prompt{
					Label:    "Enter the RHN Username",
					Validate: validate,
				}

				opt.RHNUser, err = prompt.Run()
				if err != nil {
					return err
				}
			}

			if opt.RHNPassword == "" {
				prompt := promptui.Prompt{
					Label:    "Enter the RHN Password",
					Mask:     '•',
					Validate: validate,
				}

				opt.RHNPassword, err = prompt.Run()
				if err != nil {
					return err
				}
			}
		}

		if opt.ImageDist != "coreos" && opt.OSPassword == "" && !opt.OSPasswordSkip {
			var err error
			opt.OSPassword, err = GeneratePassword(12)
			klog.Infof("Autogenerated OS root password is: %s", opt.OSPassword)
			if err != nil {
				return err
			}
		}

		// preflight checks validations
		return validate.Validate()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opt := pkg.ImageCMDOptions

		tmpDir, err := ioutil.TempDir(opt.TempDir, "qcow2ova")
		if err != nil {
			return fmt.Errorf("failed to create a temprory directory: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		mnt := filepath.Join(tmpDir, "mnt")
		err = os.Mkdir(mnt, 0755)
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		// Block for handling the interrupt and perform the cleanup
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-c
			klog.Infof("Received an interrupt, exiting...")
			_ = os.RemoveAll(tmpDir)
			prep.UmountHostPartitions(mnt)
			_ = prep.Umount(mnt)
			os.Exit(1)
		}()

		image, err := getImage(tmpDir, opt.ImageURL, 0)
		if err != nil {
			return fmt.Errorf("failed to download the %s into %s, error: %v", opt.ImageURL, tmpDir, err)
		}

		klog.Infof("downloaded/copied the file at: %s", image)

		var qcow2Img string

		checkGzip, err := isGzip(image)
		if err != nil {
			return fmt.Errorf("failed to detect the image filetype: %v", err)
		}
		if checkGzip {
			klog.Infof("Image %s is in gzip format, extracting it", image)
			qcow2Img = filepath.Join(tmpDir, ova.VolName+".qcow2")
			err = gunzipIt(image, qcow2Img)
			if err != nil {
				return err
			}
			klog.Infof("Extract complete")
		} else {
			qcow2Img = image
		}

		ovaImgDir := filepath.Join(tmpDir, "ova-img-dir")
		err = os.Mkdir(ovaImgDir, 0755)
		if err != nil {
			return err
		}

		rawImg := filepath.Join(ovaImgDir, ova.VolNameRaw)

		klog.Infof("Converting Qcow2(%s) image to raw(%s) format", qcow2Img, rawImg)
		err = qemuImgConvertQcow2Raw(qcow2Img, rawImg)
		if err != nil {
			return err
		}
		klog.Infof("Conversion completed")

		klog.Infof("Resizing the image %s to %dG", rawImg, opt.ImageSize)
		err = qemuImgResize("-f", "raw", rawImg, fmt.Sprintf("%dG", opt.ImageSize))
		if err != nil {
			return err
		}
		klog.Infof("Resize completed")

		klog.Infof("Preparing the image")
		err = prep.Prepare4capture(mnt, rawImg, opt.ImageDist, opt.RHNUser, opt.RHNPassword, opt.OSPassword)
		if err != nil {
			return fmt.Errorf("failed while preparing the image for %s distro, err: %v", opt.ImageDist, err)
		}
		klog.Infof("Preparation completed")

		klog.Infof("Creating an OVA bundle")
		ovafile := filepath.Join(tmpDir, opt.ImageName+".ova")
		if err := ova.CreateTarArchive(ovaImgDir, ovafile, opt.TargetDiskSize); err != nil {
			return fmt.Errorf("failed to create ova bundle, err: %v", err)
		}
		klog.Infof("OVA bundle creation completed: %s", ovafile)

		klog.Infof("Compressing an OVA file")
		ovaGZfile := filepath.Join(cwd, opt.ImageName+".ova.gz")
		err = gzipIt(ovafile, ovaGZfile)
		if err != nil {
			return err
		}
		klog.Infof("OVA file Compression completed")

		fmt.Printf("\n\nSuccessfully converted Qcow2 image to OVA format, find at %s\nOS root password: %s\n", ovaGZfile, opt.OSPassword)
		return nil
	},
}

func init() {
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.ImageName, "image-name", "", "Name of the resultant OVA image")
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.ImageURL, "image-url", "", "URL or absolute local file path to the <QCOW2>.gz image")
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.ImageDist, "image-dist", "", "Image Distribution(supported: rhel, centos, coreos)")
	Cmd.Flags().Uint64Var(&pkg.ImageCMDOptions.ImageSize, "image-size", 11, "Size (in GB) of the resultant OVA image")
	Cmd.Flags().Int64Var(&pkg.ImageCMDOptions.TargetDiskSize, "target-disk-size", 120, "Size (in GB) of the target disk volume where OVA will be copied")
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.RHNUser, "rhn-user", "", "RedHat Subscription username. Required when Image distribution is rhel")
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.RHNPassword, "rhn-password", "", "RedHat Subscription password. Required when Image distribution is rhel")
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.OSPassword, "os-password", "", "Root user password, will auto-generate the 12 bits password(applicable only for redhat and cento distro)")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.TempDir, "temp-dir", "t", os.TempDir(), "Scratch space to use for OVA generation")
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.PrepTemplate, "prep-template", "", "Image preparation script template, use --prep-template-default to print the default template(supported distros: rhel and centos)")
	Cmd.Flags().BoolVar(&pkg.ImageCMDOptions.PrepTemplateDefault, "prep-template-default", false, "Prints the default image preparation script template, use --prep-template to set the custom template script(supported distros: rhel and centos)")
	Cmd.Flags().StringSliceVar(&pkg.ImageCMDOptions.PreflightSkip, "skip-preflight-checks", []string{}, "Skip the preflight checks(e.g: diskspace, platform, tools) - dev-only option")
	Cmd.Flags().BoolVar(&pkg.ImageCMDOptions.OSPasswordSkip, "skip-os-password", false, "Skip the root user password")
	_ = Cmd.Flags().MarkHidden("skip-preflight-checks")
	_ = Cmd.MarkFlagRequired("image-name")
	_ = Cmd.MarkFlagRequired("image-url")
	Cmd.Flags().SortFlags = false
}
