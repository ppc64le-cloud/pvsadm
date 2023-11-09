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

package audit

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"k8s.io/klog/v2"
)

var Logger *Audit

func Log(name, op, value string) {
	Logger.Log(name, op, value)
}

type Audit struct {
	file  *os.File
	mutex *sync.Mutex
}

type log struct {
	Name      string    `json:"name"`
	Operation string    `json:"op"`
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

func New(file string) *Audit {
	logfile, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		klog.Fatalf("failed to open audit file: %s, err: %v", pkg.Options.AuditFile, err)
	}
	return &Audit{
		file:  logfile,
		mutex: &sync.Mutex{},
	}
}

func (a *Audit) Log(name, op, value string) {
	a.mutex.Lock()
	entry := log{
		Name:      name,
		Operation: op,
		Value:     value,
		Timestamp: time.Now().UTC(),
	}
	jsonEntry, err := json.Marshal(entry)
	if err != nil {
		klog.Fatalf("json marshal error: %v", err)
	}
	jsonEntry = append(jsonEntry, '\n')
	if _, err := a.file.Write(jsonEntry); err != nil {
		klog.Fatalf("log failed with error: %v", err)
	}
	a.mutex.Unlock()
}

func Delete(file string) {
	check_file, err := os.Stat(file)
	if err != nil {
		klog.V(2).Infoln(err)
		return
	}
	if check_file.Size() == 0 {
		os.Remove(file)
	}
}
