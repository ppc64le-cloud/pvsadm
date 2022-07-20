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
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	type args struct {
		dist       string
		rhnuser    string
		rhnpasswd  string
		rootpasswd string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
		notwant string
	}{
		{
			name:    "rhel image",
			args:    args{"rhel", "rhn", "rhnpassword", "some-password"},
			want:    "subscription-manager",
			wantErr: false,
		},
		{
			name:    "centos image",
			args:    args{dist: "centos", rootpasswd: "some-password"},
			want:    "some-password",
			wantErr: false,
		},
		{
			name:    "rhel image without root password",
			args:    args{dist: "rhel", rhnuser: "rhn", rhnpasswd: "rhnpassword"},
			want:    "subscription-manager",
			wantErr: false,
			notwant: "passwd root",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.args.dist, tt.args.rhnuser, tt.args.rhnpasswd, tt.args.rootpasswd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.Contains(got, tt.want) {
				t.Errorf("Render() %s does not contain the %s", got, tt.want)
			}
			if tt.notwant != "" && strings.Contains(got, tt.notwant) {
				t.Errorf("Render() %s contain the %s", got, tt.notwant)
			}
		})
	}
}
