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

package utils

import (
	"strings"
	"testing"
)

func TestRunCMD(t *testing.T) {
	type args struct {
		cmd  string
		args []string
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 string
		want2 string
	}{
		{
			"echo test",
			args{"echo", []string{"hello world"}},
			0,
			"hello world",
			"",
		},
		{
			"Command not Found",
			args{"some-command", []string{"ls"}},
			1,
			"",
			"exec: \"some-command\": executable file not found in $PATH",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := RunCMD(tt.args.cmd, tt.args.args...)
			if got != tt.want {
				t.Errorf("RunCMD() got = %v, want %v", got, tt.want)
			}
			if strings.TrimSuffix(got1, "\n") != tt.want1 {
				t.Errorf("RunCMD() got1 = %v, want %v", got1, tt.want1)
			}
			if strings.TrimSuffix(got2, "\n") != tt.want2 {
				t.Errorf("RunCMD() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
