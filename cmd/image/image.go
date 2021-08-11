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

package image

import (
	_import "github.com/ppc64le-cloud/pvsadm/cmd/image/import"
	"github.com/ppc64le-cloud/pvsadm/cmd/image/qcow2ova"
	"github.com/ppc64le-cloud/pvsadm/cmd/image/sync"
	"github.com/ppc64le-cloud/pvsadm/cmd/image/upload"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "image",
	Short: "PowerVS Image management",
	Long:  `PowerVS Image management`,
}

func init() {
	Cmd.AddCommand(_import.Cmd)
	Cmd.AddCommand(qcow2ova.Cmd)
	Cmd.AddCommand(upload.Cmd)
	Cmd.AddCommand(sync.Cmd)
}
