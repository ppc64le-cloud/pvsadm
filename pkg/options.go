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

import "time"

var Options = &options{}

type options struct {
	InstanceID   string
	APIKey       string
	Environment  string
	Region       string
	Zone         string
	DryRun       bool
	Debug        bool
	Since        time.Duration
	Before       time.Duration
	InstanceName string
	NoPrompt     bool
	IgnoreErrors bool
	AuditFile    string
	Expr         string
}

// Options for pvsadm image command
var ImageCMDOptions = &imageCMDOptions{}

type imageCMDOptions struct {
	//qcow2ova options
	ImageDist           string
	ImageName           string
	ImageSize           uint64
	TargetDiskSize      int64
	ImageURL            string
	OSPassword          string
	PreflightSkip       []string
	RHNUser             string
	RHNPassword         string
	TempDir             string
	PrepTemplate        string
	PrepTemplateDefault bool
	OSPasswordSkip      bool
	//upload options
	InstanceName string
	Region       string
	BucketName   string
	ResourceGrp  string
	ServicePlan  string
	ObjectName   string
	//import options
	COSInstanceName string
	ImageFilename   string
	AccessKey       string
	SecretKey       string
	StorageType     string
	InstanceID      string
	ServiceCredName string
	Public          bool
	Watch           bool
	WatchTimeout    time.Duration
	//sync options
	SpecYAML string
}
