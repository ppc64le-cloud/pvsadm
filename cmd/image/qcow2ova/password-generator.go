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
	"crypto/rand"
	"encoding/base64"
)

// GeneratePassword generates the password of length n
func GeneratePassword(n int) (b64Password string, err error) {
	b := make([]byte, n)
	_, err = rand.Read(b)
	if err != nil {
		return
	}
	b64Password = base64.URLEncoding.EncodeToString(b)
	return
}
