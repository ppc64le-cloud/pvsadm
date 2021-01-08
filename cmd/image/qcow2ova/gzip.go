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
	"io"
	"net/http"
	"os"
	"path/filepath"

	gzip "github.com/klauspost/pgzip"
)

// gzipIt compresses the source file to dest
func gzipIt(src, dest string) error {
	reader, err := os.Open(src)
	if err != nil {
		return err
	}

	filename := filepath.Base(src)
	writer, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	defer archiver.Close()
	archiver.Name = filename

	_, err = io.Copy(archiver, reader)
	return err
}

// gunzipIt the source file to target
func gunzipIt(src, dest string) error {
	reader, err := os.Open(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer archive.Close()

	writer, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	return err
}

// isGzip returns if file is in gzip format
func isGzip(source string) (bool, error) {
	file, err := os.Open(source)
	if err != nil {
		return false, err
	}
	defer file.Close()

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		return false, err
	}

	if filetype := http.DetectContentType(buff); filetype == "application/x-gzip" {
		return true, nil
	} else {
		return false, nil
	}
}
