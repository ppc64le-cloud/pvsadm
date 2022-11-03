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

package ova

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/version"
)

const (
	VolName    = "disk"
	VolNameRaw = VolName + ".raw"
)

type OVA struct {
	ImageName, VolumeName string
	SrcVolumeSize         int64
	TargetDiskSize        int64
	PvsadmVersion         string
	OsId                  string
}

// Render will generate the OVA spec from the template with all the required information like image name, volume name
// and size
func Render(imageName, volumeName string, srcVolumeSize int64, targetDiskSize int64) (string, error) {
	//Disk Size should be in bytes

	opt := pkg.ImageCMDOptions
	osId := "79"
	if opt.ImageDist == "coreos" {
		osId = "80"
	}
	o := OVA{
		imageName, volumeName, srcVolumeSize, targetDiskSize * 1073741824, version.Get(), osId,
	}

	var wr bytes.Buffer
	t := template.Must(template.New("ova").Parse(ovfTemplate))
	err := t.Execute(&wr, o)
	if err != nil {
		return "", fmt.Errorf("error while rendoring the ova template: %v", err)
	}
	return wr.String(), nil
}

// RenderMeta will generate the OVA meta spec from the template with the image name
func RenderMeta(imageName string) (string, error) {
	o := OVA{
		ImageName: imageName,
	}
	var wr bytes.Buffer
	t := template.Must(template.New("ova").Parse(metaTemplate))
	err := t.Execute(&wr, o)
	if err != nil {
		return "", fmt.Errorf("error while rendoring the ova template: %v", err)
	}
	return wr.String(), nil
}

// bundles the dir into a OVA image
func CreateTarArchive(dir string, target string, targetDiskSize int64) error {
	ovf := filepath.Join(dir, VolNameRaw)
	info, err := os.Stat(ovf)
	if os.IsNotExist(err) {
		return err
	}
	volSize := info.Size()
	meta, err := RenderMeta(filepath.Base(target))
	if err != nil {
		return fmt.Errorf("failed to render the meta specfile, got error '%s'", err.Error())
	}
	ovfSpec, err := Render(filepath.Base(target), VolNameRaw, volSize, targetDiskSize)
	if err != nil {
		return fmt.Errorf("failed to render the ovf specfile, got error '%s'", err.Error())
	}

	file, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("failed to create file '%s', got error '%s'", target, err.Error())
	}
	defer file.Close()
	tw := tar.NewWriter(file)

	defer tw.Close()

	// Write the ovf and meta files
	var files = []struct {
		Name, Body string
	}{
		{"coreos.ovf", ovfSpec},
		{"coreos.meta", meta},
	}
	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return fmt.Errorf("could not write header for file '%s', got error '%s'", file, err.Error())
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			return fmt.Errorf("could not write the content for file '%s', got error '%s'", file, err.Error())
		}
	}

	// include the ovf volume
	hrd := &tar.Header{
		Name:    filepath.Base(ovf),
		Size:    volSize,
		Mode:    int64(info.Mode()),
		ModTime: info.ModTime(),
	}

	err = tw.WriteHeader(hrd)
	if err != nil {
		return fmt.Errorf("could not write header for file '%s', got error '%s'", ovf, err.Error())
	}

	ovfFD, err := os.Open(ovf)
	if err != nil {
		return fmt.Errorf("failed to open a ovf file: %s", ovf)
	}
	defer ovfFD.Close()

	_, err = io.Copy(tw, ovfFD)
	if err != nil {
		return fmt.Errorf("could not copy the file '%s' data to the tarball, got error '%s'", ovf, err.Error())
	}

	return nil
}
