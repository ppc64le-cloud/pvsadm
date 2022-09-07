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
	"archive/tar"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

// gunzipIt the source file to target
func gunzipIt(src, dest string) error {
	reader, err := os.Open(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer archive.Close()

	writer, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	return err
}

// isGzip returns if file is in gzip format
func isGzip(source string) (bool, error) {
	file, err := os.Open(source)
	if err != nil {
		return false, err
	}
	defer file.Close()

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		return false, err
	}

	if filetype := http.DetectContentType(buff); filetype == "application/x-gzip" {
		return true, nil
	} else {
		return false, nil
	}
}

func SanitizeExtractPath(filePath string, destination string) error {
	destpath := filepath.Join(destination, filePath)
	if !strings.HasPrefix(destpath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("%s: illegal file path", filePath)
	}
	return nil
}

// Extract specific file from the tar file
func Untar(tarball, target, filename string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if header.Name != filename {
			continue
		}
		err = SanitizeExtractPath(header.Name, target)
		if err != nil {
			return err
		}
		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}

var Cmd = &cobra.Command{
	Use:   "info",
	Short: "Provide info about the image",
	Long: `Provide info about the image
pvsadm image info --help for information


Examples:

# To get the pvsadm tool version by specifying the image
pvsadm image info --file rhcos-46-12152021.ova.gz

`,
	RunE: func(cmd *cobra.Command, args []string) error {

		opt := pkg.ImageCMDOptions
		var ova string
		ovaImgDir, err := os.MkdirTemp(opt.TemDir, "ova-img-dir")
		if err != nil {
			return err
		}

		defer os.RemoveAll(ovaImgDir)

		//Check if the image is in gzip format and unzip it.
		checkGzip, err := isGzip(opt.Filename)
		if err != nil {
			return fmt.Errorf("failed to detect the image filetype: %v", err)
		}
		if checkGzip {
			klog.Infof("Image %s is in gzip format, extracting it", opt.Filename)
			ova = filepath.Join(ovaImgDir, "image.ova")
			err = gunzipIt(opt.Filename, ova)
			if err != nil {
				return err
			}
			klog.Infof("Extract complete")
		} else {
			ova = opt.Filename
		}

		//Extract the ovf file
		err = Untar(ova, ovaImgDir, "coreos.ovf")
		if err != nil {
			return err
		}
		filename := filepath.Join(ovaImgDir, "coreos.ovf")

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
		byteValue, _ := ioutil.ReadAll(xmlFile)
		xml.Unmarshal(byteValue, &version)
		toolversion := version.Pvsadmversionsection.BuildNumber
		if toolversion == "" {
			return fmt.Errorf("unable to find the pvsadm tool version")
		}
		klog.Infof("Pvsadm tool version used for creating this image: %+v", toolversion)
		return nil
	},
}

// Init method
func init() {
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.Filename, "file", "f", "", "The PATH to the image")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.TemDir, "temp-dir", "t", os.TempDir(), "Scratch space to use for OVA extraction")
	_ = Cmd.MarkFlagRequired("file")
	Cmd.Flags().SortFlags = false
}
