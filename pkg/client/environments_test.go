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

package client

import (
	"reflect"
	"testing"
)

func TestListEnvironments(t *testing.T) {
	tests := []struct {
		name     string
		wantKeys []string
	}{
		{
			"valid environments",
			[]string{"test", "prod"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotKeys := ListEnvironments(); !reflect.DeepEqual(gotKeys, tt.wantKeys) {
				t.Errorf("ListEnvironments() = %v, want %v", gotKeys, tt.wantKeys)
			}
		})
	}
}

func TestGetEnvironment(t *testing.T) {
	type args struct {
		env string
	}
	tests := []struct {
		name  string
		args  args
		want1 map[string]string
		want  error
	}{
		{
			"valid environment - test",
			args{"test"},
			Environments["test"],
			nil,
		},
		{
			"valid environment - prod",
			args{"prod"},
			Environments["prod"],
			nil,
		},
		{
			"valid environment - prod",
			args{"fake"},
			nil,
			ErrEnvironmentNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1, got := GetEnvironment(tt.args.env)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEnvironment() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetEnvironment() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
