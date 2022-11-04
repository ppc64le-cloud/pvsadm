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

package utils

import (
	"testing"
)

func TestSanitizeExtractPath(t *testing.T) {
	type args struct {
		filePath    string
		destination string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"with proper filepath",
			args{"/test", "new"},
			false,
		},
		{
			"With improper filepath",
			args{"../test", "new"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeExtractPath(tt.args.filePath, tt.args.destination)
			if (got != nil) != tt.wantErr {
				t.Errorf("SanitizeExtractPath() got  %v, wantErr %v", got, tt.wantErr)
			}
		})
	}

}
