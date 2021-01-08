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

package pkg

import (
	"testing"
	"time"
)

func TestIsPurgeable(t *testing.T) {
	type args struct {
		candidate time.Time
		before    time.Duration
		since     time.Duration
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Purgeable if both before and since are not set",
			args{time.Now(), 0, 0},
			true,
		},
		{
			"Both before and since aren't supported at a time",
			args{time.Now(), 1 * time.Minute, 1 * time.Minute},
			false,
		},
		{
			"Purgeable candidate before",
			args{time.Now().Add(-10 * time.Hour), 9 * time.Hour, 0},
			true,
		},
		{
			"non-Purgeable candidate before",
			args{time.Now().Add(-10 * time.Hour), 11 * time.Hour, 0},
			false,
		},
		{
			"Purgeable candidate since",
			args{time.Now().Add(-10 * time.Hour), 0, 11 * time.Hour},
			true,
		},
		{
			"non-Purgeable candidate since",
			args{time.Now().Add(-10 * time.Hour), 0, 9 * time.Hour},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPurgeable(tt.args.candidate, tt.args.before, tt.args.since)
			if got != tt.want {
				t.Errorf("IsPurgeable() got = %v, want %v", got, tt.want)
			}
		})
	}
}
