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
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_getImage(t *testing.T) {
	// HTTP server serving the request with processing time atleast 2 seconds
	httpProcessingTime := 2 * time.Second
	mux := http.NewServeMux()
	mux.HandleFunc("/file", func(w http.ResponseWriter, req *http.Request) {
		time.Sleep(httpProcessingTime)
		fmt.Fprintf(w, "Hello World!")
	})
	mux.HandleFunc("/file1", func(w http.ResponseWriter, req *http.Request) {
		time.Sleep(httpProcessingTime)
		fmt.Fprintf(w, "Hello World!")
	})
	mux.HandleFunc("/file2", func(w http.ResponseWriter, req *http.Request) {
		time.Sleep(httpProcessingTime)
		fmt.Fprintf(w, "Hello World!")
	})
	mux.HandleFunc("/fail", func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, fmt.Sprintf("failed to handle the build"), http.StatusInternalServerError)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	content := []byte("temporary file's content")
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	destDir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.Fatal(err)
	}
	//defer os.RemoveAll(destDir)

	tmpfn := filepath.Join(dir, "tmpfile")
	if err := ioutil.WriteFile(tmpfn, content, 0666); err != nil {
		log.Fatal(err)
	}

	type args struct {
		dir     string
		src     string
		timeout time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "getImage of type file",
			args:    args{destDir, tmpfn, 0},
			want:    filepath.Join(destDir, "tmpfile"),
			wantErr: false,
		},
		{
			name:    "getImage does not exist",
			args:    args{destDir, "/file/doesnot/exist", 0},
			want:    "",
			wantErr: true,
		},
		{
			name:    "getImage of type URL",
			args:    args{destDir, ts.URL + "/file1", httpProcessingTime * 2},
			want:    filepath.Join(destDir, "file1"),
			wantErr: false,
		},
		{
			name:    "getImage of type URL with default timeout",
			args:    args{destDir, ts.URL + "/file2", 0},
			want:    filepath.Join(destDir, "file2"),
			wantErr: false,
		},
		{
			name:    "getImage of type URL - timeout failure",
			args:    args{destDir, ts.URL + "/file", httpProcessingTime / 2},
			want:    "",
			wantErr: true,
		},
		{
			name:    "getImage of type URL - server side error",
			args:    args{destDir, ts.URL + "/fail", 0},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getImage(tt.args.dir, tt.args.src, tt.args.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("getImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getImage() got = %v, want %v", got, tt.want)
			}
		})
	}
}
