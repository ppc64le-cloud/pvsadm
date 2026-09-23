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
	"crypto/sha256"
	"encoding/hex"
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

// verifyCheckSum validates SHA256 of a downloaded file
func verifyCheckSum(filePath, expected string) error {
	if expected == "" {
		klog.V(1).Infof("No checksum provided for %s, skipping verification", filePath)
		return nil
	}
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for checksum: %v", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("failed to calculate checksum: %v", err)
	}

	actual := hex.EncodeToString(h.Sum(nil))
	if actual != expected {
		return fmt.Errorf("checksum mismatch for %s:\n expected: %s\n actual:	%s", filePath, expected, actual)
	}
	klog.V(1).Infof("Checksum verification PASSED FOR %s", filePath)
	return nil
}

// Downloads or copy the image into the target dir mentioned
// Added checksum verification (optional)
func getImage(downloadDir string, srcUrl string, timeout time.Duration, expectedSha string) (string, error) {
	if timeout == 0 {
		timeout = DefaultGetTimeout
	}
	dest := path.Join(downloadDir, path.Base(srcUrl))
	if !isURL(srcUrl) {
		if !fileExists(srcUrl) {
			return "", fmt.Errorf("not a valid URL or file does not exist at %s", srcUrl)
		}
		klog.V(1).Infof("Copying %s into %s", srcUrl, dest)
		if err := cp(srcUrl, dest); err != nil {
			return "", err
		}
		klog.V(1).Info("Copy Completed!")
	} else {
		out, err := os.Create(dest)
		if err != nil {
			return "", err
		}
		defer out.Close()
		klog.V(1).Infof("Downloading %s into %s", srcUrl, dest)
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
		klog.V(1).Info("Download Completed!")
	}
	// Verify checksum if provided
	if err := verifyCheckSum(dest, expectedSha); err != nil {
		return "", err
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
