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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"k8s.io/klog/v2"
)

const (
	DefaultGetTimeout = 30 * time.Minute
)

// Downloads or copy the image into the target dir mentioned
func getImage(downloadDir string, srcUrl string, timeout time.Duration) (string, error) {
	if timeout == 0 {
		timeout = DefaultGetTimeout
	}
	dest := path.Join(downloadDir, path.Base(srcUrl))
	if !isURL(srcUrl) {
		if !fileExists(srcUrl) {
			return "", fmt.Errorf("not a valid URL or file does not exist at %s", srcUrl)
		}
		klog.Infof("Copying %s into %s", srcUrl, dest)
		if err := cp(srcUrl, dest); err != nil {
			return "", err
		}
		klog.Infof("Copy Completed!")
	} else {
		out, err := os.Create(dest)
		if err != nil {
			return "", err
		}
		defer out.Close()
		klog.Infof("Downloading %s into %s", srcUrl, dest)
		client := http.Client{
			Timeout: timeout,
		}
		resp, err := client.Get(srcUrl)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to download the file: %s, status code: %d", srcUrl, resp.StatusCode)
		}

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return "", err
		}
		klog.Infof("Download Completed!")
	}
	return dest, nil
}

// Copies the src to dest
func cp(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func isURL(URL string) bool {
	_, err := url.ParseRequestURI(URL)
	if err != nil {
		return false
	}
	u, err := url.Parse(URL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

// fileExists check if file exists or not and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
