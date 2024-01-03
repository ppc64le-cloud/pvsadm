// Copyright 2022 IBM Corp
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

package info

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var Cmd = &cobra.Command{
	Use:   "info",
	Short: "Provide info about the image",
	Long: `Provide info about the image
pvsadm image info --help for information

Examples:
# To get the pvsadm tool version by specifying the image
pvsadm image info rhcos-46-12152021.ova.gz

`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) == 0 {
			return fmt.Errorf("Image filepath is required")
		}

		fileName := args[0]
		var ova string
		ovaImgDir, err := os.MkdirTemp(os.TempDir(), "ova-img-dir")
		if err != nil {
			return err
		}

		defer os.RemoveAll(ovaImgDir)

		//Check if the image is in gzip format and unzip it.
		checkGzip, err := utils.IsGzip(fileName)
		if err != nil {
			return fmt.Errorf("failed to detect the image filetype: %v", err)
		}
		if checkGzip {
			klog.Infof("Image %s is in gzip format, extracting it", fileName)
			ova = filepath.Join(ovaImgDir, "image.ova")
			err = utils.GunzipIt(fileName, ova)
			if err != nil {
				return err
			}
			klog.Infof("Extract complete")
		} else {
			ova = fileName
		}

		//Extract the ovf file
		err = utils.Untar(ova, ovaImgDir, "*.ovf")
		if err != nil {
			return err
		}
		ovfFilePath := ovaImgDir + "/*.ovf"
		matches, err := filepath.Glob(ovfFilePath)
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			return fmt.Errorf(".ovf file is not found in %s image", fileName)
		}
		filename := matches[0]

		type Ovf struct {
			Pvsadmversionsection struct {
				Info        string `xml:"Info"`
				BuildNumber string `xml:"BuildNumber"`
			} `xml:"pvsadmversionsection"`
		}

		xmlFile, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer xmlFile.Close()
		var version Ovf
		byteValue, err := io.ReadAll(xmlFile)
		if err != nil {
			return err
		}
		err = xml.Unmarshal(byteValue, &version)
		if err != nil {
			return err
		}
		toolversion := version.Pvsadmversionsection.BuildNumber
		if toolversion == "" {
			klog.Warning("unable to find the pvsadm tool version")
			return nil
		}
		klog.Infof("Pvsadm tool version used for creating this image: %+v", toolversion)
		return nil
	},
}
