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

// Specifications
type Spec struct {
	Source `yaml:"source"`
	Target []TargetItem `yaml:"target"`
}

// Source Specifications
type Source struct {
	Bucket       string `yaml:"bucket"`
	Cos          string `yaml:"cos"`
	Object       string `yaml:"object"`
	StorageClass string `yaml:"storageClass"`
	Region       string `yaml:"region"`
}

// TargetItem Specifications
type TargetItem struct {
	Bucket       string `yaml:"bucket"`
	StorageClass string `yaml:"storageClass"`
	Region       string `yaml:"region"`
}
