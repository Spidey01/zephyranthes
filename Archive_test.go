// SPDX-License-Identifier: Zlib
// Copyright 2026, Terry M. Poulin.

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// This file provides functionality for testing Archive implementations.

// Set this to use a path to use instead of [os.MkdirTemp] for creating test
// inputs via [MakeTestDir].
var forceTestdata string

// This makes the cleanup function from [MakeTestDir] a no-op.
var keepTestDir bool

type ExtractableArchive interface {
	Archive
	// Extracts the contents of the archive into the specified root directory.
	// This is not part of Zephyranthes' normal functionality, but is required
	// for unit test coverage.
	ExtractTo(root string) error
}

// Create a temporary directory and return it along with a clean up function.
func MakeTestDir(t *testing.T) (string, func()) {
	var dir string
	var err error
	if forceTestdata != "" {
		if err = os.MkdirAll(forceTestdata, os.ModePerm); err != nil {
			t.Fatalf("os.MkdirAll(%q, ...) failed: %v", forceTestdata, err)
		}
		dir = forceTestdata
	} else {
		dir, err = os.MkdirTemp("", "")
		if err != nil {
			t.Fatalf("MakeTestDir(t) failed: %v", err)
		}
	}
	cleanup := func() {
		if forceTestdata != "" {
			t.Logf("Skipping %q cleanup because forceTestdata is set", forceTestdata)
			return
		}
		if keepTestDir {
			t.Logf("Skipping %q cleanup because keepTestDir is set", dir)
			return
		}
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("os.RemoveAll(%q) failed: %v", dir, err)
		} else {
			t.Logf("os.RemoveAll(%q)", dir)
		}
	}
	t.Logf("MakeTestDir(t) => %q, cleanup", dir)
	return dir, cleanup
}

// Defines test data to use as an input for archiving.
type TestInput struct {
	Path     string      // Actual path on dist.
	Name     string      // Name to put in archive.
	File     *os.File    // File handle to add, must close if non-nil.
	DirEntry os.DirEntry // Directory handle to add.
	Stat     fs.FileInfo // Info on the File or DirEntry.
	Contents []byte      // Contains a copy of the file contents when File is non-nil.
}

func getDirEntry(path string) (fs.DirEntry, error) {
	prefix := filepath.Dir(path)
	base := filepath.Base(path)
	dirEntries, err := os.ReadDir(prefix)
	if err != nil {
		return nil, fmt.Errorf("os.ReadDir(%q) failed: %v", prefix, err)
	}
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() && dirEntry.Name() == base {
			return dirEntry, nil
		}
	}
	return nil, fmt.Errorf("no fs.DirEntry found for %q", path)
}

// Creates the specified directory and returns a TestInput for it.
func MkdirTestInput(t *testing.T, path, name string) TestInput {
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatalf("MkdirTestInput(...): os.Mkdir(%q): failed: %v", path, err)
	}
	dirEntry, err := getDirEntry(path)
	if err != nil {
		t.Fatalf("getDirEntry(%q) failed: %v", path, err)
	}
	stat, err := dirEntry.Info()
	if err != nil {
		t.Fatalf("MkdirTestInput(...): dirEntry.Info() failed: %v", err)
	}
	return TestInput{
		Path:     path,
		Name:     name,
		DirEntry: dirEntry,
		Stat:     stat,
	}
}

// Creates the specified file and returns a TestInput for it. If contents is
// non-nil, it will be used as the file's contents. The caller is responsible
// for closing the file handle.
func MkfileTestInput(t *testing.T, path, name string, contents []byte) TestInput {
	fp, err := os.Create(path)
	if err != nil {
		t.Fatalf("MkfileTestInput(...): os.Create(%q) failed: %v", path, err)
	}
	if contents != nil {
		if _, err = fp.Write(contents); err != nil {
			t.Fatalf("MkfileTestInput(...): fp.Write() failed: %v", err)
		}
		fp.Seek(0, 0)
	}
	// Need to stat after I/O, or size will be wrong for check().
	stat, err := fp.Stat()
	if err != nil {
		t.Fatalf("MkfileTestInput(...): fp.Stat() failed: %v", err)
	}
	return TestInput{
		Path:     path,
		Name:     name,
		File:     fp,
		Stat:     stat,
		Contents: contents,
	}
}

// Creates a symlink and returns a TestInput for it. This ensures the file
// exists. Actually verifying in check(), not so much since not all formats
// support storing symlinks.
func MklinkTestInput(t *testing.T, oldpath, newpath, name string) TestInput {
	if err := os.Symlink(oldpath, newpath); err != nil {
		t.Errorf("MklinkTestInput(...): os.Link(%q, %q): failed: %v", oldpath, newpath, err)
	}
	fp, err := os.Open(newpath)
	if err != nil {
		t.Fatalf("MklinkTestInput(...): os.Open(%q) failed: %v", newpath, err)
	}
	stat, err := fp.Stat()
	if err != nil {
		t.Fatalf("MklinkTestInput(...): fp.Stat() failed: %v", err)
	}
	if !stat.IsDir() {
		return TestInput{
			Path: newpath,
			Name: name,
			File: fp,
			Stat: stat,
		}
	}
	// then need to make a directory style TestInput
	fp.Close()
	dirEntry, err := getDirEntry(oldpath)
	if err != nil {
		t.Fatalf("MklinkTestInput(...): getDirEntry(%q) failed: %v", oldpath, err)
	}
	stat, err = dirEntry.Info()
	if err != nil {
		t.Fatalf("MklinkTestInput(...): dirEntry.Info() failed: %v", err)
	}
	return TestInput{
		Path:     newpath,
		Name:     name,
		DirEntry: dirEntry,
		Stat:     stat,
	}
}

// Checks if the input was extracted into root. Returns an error if not, or a failure occurs.
func (input *TestInput) check(root string) error {
	path := filepath.Join(root, input.Name)
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("os.Stat(%q): %v", path, err)
	}
	if stat.IsDir() != input.Stat.IsDir() {
		return fmt.Errorf("name: %q actual IsDir: %v expected IsDir: %v", input.Name, stat.IsDir(), input.Stat.IsDir())
	}
	if stat.IsDir() {
		return nil
	}
	if stat.Size() != input.Stat.Size() {
		return fmt.Errorf("name: %q actual Size: %v expected Size: %v", input.Name, stat.Size(), input.Stat.Size())
	}
	// Verify the file contents.
	if input.Contents != nil {
		contents, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("os.ReadFile(%q) failed: %v", path, err)
		}
		if !slices.Equal(contents, input.Contents) {
			// TODO: hex dump both.
			return fmt.Errorf("file contents of %q differ", input.Name)
		}
	}
	return nil
}

// Creates test data in `root` and returns the list of TestInput structures
// describing it. Test fixtures include various test cases:
//
// - empty directories and subdirectories with contents
// - files with normal content and sparse files
// - symbolic links across the directory hierachy.
func MakeTestData(t *testing.T, root string) []TestInput {
	var contents []TestInput

	// Create an empty directory.
	contents = append(contents, MkdirTestInput(t, filepath.Join(root, "emptydir"), "emptydir"))

	// Create a file with many zeros, like `dd if=/dev/zero of=zeros.dd bs=1
	// count=5000`. Because of how spare files are handled, the best way is just
	// to use the truncate system call. That means we also need to Stat it again.
	zeros := MkfileTestInput(t, filepath.Join(root, "zeros.dd"), "zeros.dd", nil)
	zeros.File.Truncate(5000)
	var err error
	if zeros.Stat, err = zeros.File.Stat(); err != nil {
		t.Fatalf("zeros.File.Stat() failed: %v", err)
	}
	contents = append(contents, zeros)

	// Create a subdir with a text file and a symlink to the file above.
	subdir := MkdirTestInput(t, filepath.Join(root, "subdir"), "subdir")
	contents = append(contents, subdir)
	file := MkfileTestInput(t, filepath.Join(subdir.Path, "file.txt"), filepath.Join(subdir.Name, "file.txt"), []byte("a file in a subdir\n"))
	contents = append(contents, file)

	linkup := MklinkTestInput(t, zeros.Path, filepath.Join(subdir.Path, "linkup"), filepath.Join(subdir.Name, "linkup"))
	dirlinkup := MklinkTestInput(t, subdir.Path, filepath.Join(subdir.Path, "dirlinkup"), filepath.Join(subdir.Name, "dirlinkup"))
	contents = append(contents, linkup, dirlinkup)

	return contents
}

// Populates the archive with the test data from contents. The caller must close
// the archive to finalize it.
func buildArchive(t *testing.T, archive Archive, contents []TestInput) {
	for _, input := range contents {
		if input.File != nil {
			input.File.Seek(0, 0)
			if err := archive.AddFile(input.File, input.Stat, input.Name); err != nil {
				t.Errorf("archive.AddFile failed: %v", err)
			}
			input.File.Close()
		} else if input.DirEntry != nil {
			if err := archive.AddDir(input.DirEntry, input.Stat, input.Name); err != nil {
				t.Errorf("archive.AddDir failed: %v", err)
			}
		}
	}
	if err := archive.Flush(); err != nil {
		t.Errorf("archive.Flush() failed: %v", err)
	}
}

func assertArchive(t *testing.T, format string) {
	// Create a new temporary directory for the test output.
	tmpdir, cleanup := MakeTestDir(t)
	defer cleanup()

	// Create a new archive from the factory.
	path := filepath.Join(tmpdir, "archive."+format)
	realArchive, err := CreateArchive(path, format)
	if err != nil {
		t.Fatalf("CreateArchive(%q, %q) failed: %v", path, format, err)
	}
	archive, ok := realArchive.(ExtractableArchive)
	if !ok {
		t.Fatalf("Archive implementation for %q format must implement ExtractableArchive to support unit tests", format)
	}
	if archive.Name() != path {
		t.Errorf("%s Archive.Name(): returned %q but %q was expected", format, archive.Name(), path)
	}

	// Create the test fixtures and build the archive.
	testdataDir, cleanupTestData := MakeTestDir(t)
	defer cleanupTestData()
	testdata := MakeTestData(t, testdataDir)
	buildArchive(t, archive, testdata)
	// We need to close the archive to ensure it's written out.
	if err = archive.Close(); err != nil {
		t.Fatalf("%s Archive.Close(): failed: %v", format, err)
	}

	// Ensure we can build the archive with the test data and that we can
	// extract it accordingly.
	// assertBuildArchive(t, archive)
	if err = archive.ExtractTo(tmpdir); err != nil {
		t.Fatalf("%s Archive.ExtractTo(%q) failed: %v", format, tmpdir, err)
	}
	for _, input := range testdata {
		if err = input.check(tmpdir); err != nil {
			t.Errorf("%s extraction check failed: %v", format, err)
		}
	}
}

func TestArchive(t *testing.T) {
	t.Run("zip", func(t *testing.T) {
		assertArchive(t, FormatZip)
	})
	t.Run("tar", func(t *testing.T) {
		assertArchive(t, FormatTar)
	})
	// These require a decompression filter
	t.Run("tgz", func(t *testing.T) {
		assertArchive(t, FormatTarGz)
		assertArchive(t, FormatTGZ)
	})
}
