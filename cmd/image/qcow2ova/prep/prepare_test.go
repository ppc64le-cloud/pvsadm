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

package prep

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── copyFiles / copyFile / copyDir ────────────────────────────────────────────

func TestCopyFile_Basic(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.txt")
	dst := filepath.Join(tmp, "dst.txt")

	if err := os.WriteFile(src, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyFiles(src, dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Errorf("content mismatch: got %q", got)
	}
}

func TestCopyFile_PreservesPermissions(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "exec.sh")
	dst := filepath.Join(tmp, "exec_copy.sh")

	if err := os.WriteFile(src, []byte("#!/bin/sh"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := copyFiles(src, dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode() != 0755 {
		t.Errorf("expected mode 0755, got %v", info.Mode())
	}
}

func TestCopyFiles_RejectsSymlink(t *testing.T) {
	tmp := t.TempDir()
	real := filepath.Join(tmp, "real.txt")
	link := filepath.Join(tmp, "link.txt")

	if err := os.WriteFile(real, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}

	err := copyFiles(link, filepath.Join(tmp, "dst.txt"))
	if err == nil {
		t.Fatal("expected error for symlink, got nil")
	}
	if !strings.Contains(err.Error(), "symlinks are not supported") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCopyDir_Basic(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "srcdir")
	dst := filepath.Join(tmp, "dstdir")

	if err := os.Mkdir(src, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte("aaa"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "b.txt"), []byte("bbb"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyFiles(src, dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, name := range []string{"a.txt", "b.txt"} {
		data, err := os.ReadFile(filepath.Join(dst, name))
		if err != nil {
			t.Fatalf("missing file %s: %v", name, err)
		}
		if name == "a.txt" && string(data) != "aaa" {
			t.Errorf("a.txt: got %q", data)
		}
		if name == "b.txt" && string(data) != "bbb" {
			t.Errorf("b.txt: got %q", data)
		}
	}
}

func TestCopyDir_Nested(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "srcdir")
	sub := filepath.Join(src, "sub")
	dst := filepath.Join(tmp, "dstdir")

	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "nested.txt"), []byte("nested"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyFiles(src, dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "sub", "nested.txt"))
	if err != nil {
		t.Fatalf("missing nested file: %v", err)
	}
	if string(data) != "nested" {
		t.Errorf("nested.txt: got %q", data)
	}
}

func TestCopyDir_RejectsSymlinkInsideDir(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "srcdir")
	real := filepath.Join(tmp, "real.txt")
	link := filepath.Join(src, "link.txt")
	dst := filepath.Join(tmp, "dstdir")

	if err := os.Mkdir(src, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(real, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}

	err := copyFiles(src, dst)
	if err == nil {
		t.Fatal("expected error for symlink inside dir, got nil")
	}
	if !strings.Contains(err.Error(), "symlinks are not supported") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// path-escape guard (the logic extracted from prepare())
// checkWritePath mirrors the guard in prepare() so we can unit-test it
// without needing a mounted image.
func checkWritePath(mnt, writeToDirPath string) error {
	imageWritePath := filepath.Join(mnt, writeToDirPath)
	cleanMnt := filepath.Clean(mnt)
	cleanImageWritePath := filepath.Clean(imageWritePath)
	if cleanImageWritePath != cleanMnt && !strings.HasPrefix(cleanImageWritePath, cleanMnt+"/") {
		return fmt.Errorf("write-to-dir-path %q escapes the image mount root", writeToDirPath)
	}
	return nil
}

func TestPathEscapeGuard(t *testing.T) {
	mnt := "/tmp/qcow2ovaXXX/mnt"

	cases := []struct {
		name        string
		path        string
		expectError bool
	}{
		{"normal absolute path", "/home/user", false},
		{"normal relative path", "home/user", false},
		{"root of mount", "/", false},
		{"dot (mnt itself)", ".", false},
		{"escape with ..", "../../etc", true},
		{"escape with leading slash+..", "/../../etc", true},
		{"escape to parent", "../", true},
		{"escape sibling dir", "../sibling", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkWritePath(mnt, tc.path)
			if tc.expectError && err == nil {
				t.Errorf("expected escape error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// copyFiles edge cases

func TestCopyFiles_NonExistentSource(t *testing.T) {
	tmp := t.TempDir()
	err := copyFiles(filepath.Join(tmp, "does_not_exist.txt"), filepath.Join(tmp, "dst.txt"))
	if err == nil {
		t.Fatal("expected error for missing source, got nil")
	}
}

func TestCopyFile_OverwritesExisting(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.txt")
	dst := filepath.Join(tmp, "dst.txt")

	// pre-populate destination with different content
	if err := os.WriteFile(dst, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyFiles(src, dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Errorf("expected truncated content %q, got %q", "new", got)
	}
}

func TestCopyDir_Empty(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "empty")
	dst := filepath.Join(tmp, "emptydst")

	if err := os.Mkdir(src, 0755); err != nil {
		t.Fatal(err)
	}

	if err := copyFiles(src, dst); err != nil {
		t.Fatalf("unexpected error copying empty dir: %v", err)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("destination dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected destination to be a directory")
	}
}

func TestCopyDir_PreservesPermissions(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "srcdir")
	dst := filepath.Join(tmp, "dstdir")

	if err := os.Mkdir(src, 0750); err != nil {
		t.Fatal(err)
	}
	// Read back the actual mode after umask is applied, so the assertion is
	// not sensitive to the host umask setting.
	srcInfo, err := os.Stat(src)
	if err != nil {
		t.Fatal(err)
	}

	if err := copyFiles(src, dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dstInfo, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if dstInfo.Mode().Perm() != srcInfo.Mode().Perm() {
		t.Errorf("expected dir mode %v, got %v", srcInfo.Mode().Perm(), dstInfo.Mode().Perm())
	}
}

// writeFilesList integration: base-name stripping and empty-list no-op
// writeFilesToImage mirrors the file-injection block in prepare() so it can be
// tested without a mounted image or chroot.
func writeFilesToImage(mnt, writeToDirPath string, writeFilesList []string) error {
	if len(writeFilesList) == 0 {
		return nil
	}
	imageWritePath := filepath.Join(mnt, writeToDirPath)
	cleanMnt := filepath.Clean(mnt)
	cleanImageWritePath := filepath.Clean(imageWritePath)
	if cleanImageWritePath != cleanMnt && !strings.HasPrefix(cleanImageWritePath, cleanMnt+"/") {
		return fmt.Errorf("write-to-dir-path %q escapes the image mount root", writeToDirPath)
	}
	if err := os.MkdirAll(imageWritePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", imageWritePath, err)
	}
	for _, filePath := range writeFilesList {
		base := filepath.Base(filePath)
		destinationPath := filepath.Join(imageWritePath, base)
		if err := copyFiles(filePath, destinationPath); err != nil {
			return err
		}
	}
	return nil
}

func TestWriteFilesToImage_CopiesFiles(t *testing.T) {
	tmp := t.TempDir()
	mnt := filepath.Join(tmp, "mnt")
	if err := os.Mkdir(mnt, 0755); err != nil {
		t.Fatal(err)
	}

	// source files in a separate directory
	srcDir := filepath.Join(tmp, "src")
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	f1 := filepath.Join(srcDir, "file1.txt")
	f2 := filepath.Join(srcDir, "file2.txt")
	if err := os.WriteFile(f1, []byte("one"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f2, []byte("two"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := writeFilesToImage(mnt, "/home/user", []string{f1, f2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for name, want := range map[string]string{"file1.txt": "one", "file2.txt": "two"} {
		got, err := os.ReadFile(filepath.Join(mnt, "home", "user", name))
		if err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
		if string(got) != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
	}
}

func TestWriteFilesToImage_BaseNameStripping(t *testing.T) {
	// A file at /deep/nested/path/report.log should land as report.log, not
	// reproduce the full source path inside the image.
	tmp := t.TempDir()
	mnt := filepath.Join(tmp, "mnt")
	if err := os.MkdirAll(mnt, 0755); err != nil {
		t.Fatal(err)
	}

	deep := filepath.Join(tmp, "a", "b", "c")
	if err := os.MkdirAll(deep, 0755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(deep, "report.log")
	if err := os.WriteFile(src, []byte("log"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := writeFilesToImage(mnt, "/logs", []string{src}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only the base name should exist at the target
	dest := filepath.Join(mnt, "logs", "report.log")
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("expected %s to exist: %v", dest, err)
	}
}

func TestWriteFilesToImage_EmptyListIsNoop(t *testing.T) {
	tmp := t.TempDir()
	mnt := filepath.Join(tmp, "mnt")
	if err := os.Mkdir(mnt, 0755); err != nil {
		t.Fatal(err)
	}

	if err := writeFilesToImage(mnt, "/home/user", []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The target directory must NOT have been created
	entries, err := os.ReadDir(mnt)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected mnt to be empty, got %d entries", len(entries))
	}
}

func TestWriteFilesToImage_EscapingPathRejected(t *testing.T) {
	tmp := t.TempDir()
	mnt := filepath.Join(tmp, "mnt")
	if err := os.Mkdir(mnt, 0755); err != nil {
		t.Fatal(err)
	}

	// empty list short-circuits before the guard — use a non-empty list
	src := filepath.Join(tmp, "x.txt")
	if err := os.WriteFile(src, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	err := writeFilesToImage(mnt, "../../etc", []string{src})
	if err == nil {
		t.Fatal("expected escape error, got nil")
	}
	if !strings.Contains(err.Error(), "escapes the image mount root") {
		t.Errorf("unexpected error: %v", err)
	}
}

// Prepare4capture distro routing
func TestPrepare4capture_UnsupportedDistro(t *testing.T) {
	err := Prepare4capture("/mnt", "/dev/null", "ubuntu", "", "", "", "", nil)
	if err == nil {
		t.Fatal("expected error for unsupported distro, got nil")
	}
	if !strings.Contains(err.Error(), "not a supported distro") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPrepare4capture_DistroIsCaseInsensitive(t *testing.T) {
	// "COREOS" (uppercase) must be treated identically to "coreos" — it should
	// return nil (no-op) rather than "not a supported distro".
	err := Prepare4capture("/mnt", "/dev/null", "COREOS", "", "", "", "", nil)
	if err != nil {
		t.Errorf("expected nil for coreos (case-insensitive), got: %v", err)
	}
}
