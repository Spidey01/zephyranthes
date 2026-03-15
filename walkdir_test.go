// SPDX-License-Identifier: Zlib
// Copyright 2026, Terry M. Poulin.

package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

// type WalkDirFunc func(path string, d DirEntry, err error) error
// func walkDir(name string, d fs.DirEntry, walkDirFn fs.WalkDirFunc) error

func TestWalkDir(t *testing.T) {
	root, cleanup := MakeTestDir(t)
	defer cleanup()
	inputs := MakeTestData(t, root)
	t.Run("badroot", func(t *testing.T) {
		root := "/does/not/exist"
		fn := func(path string, d fs.DirEntry, err error) error {
			if path != root {
				t.Errorf("WalkDir(%q, ...) called function with wonky path %q", root, path)
			}
			if d != nil {
				t.Errorf("WalkDir() called function with a fs.DirEntry, but root does not exist!")
			}
			if err == nil {
				t.Errorf("WalkDir() called function without error, but root does not exist!")
			}
			return err
		}
		if err := WalkDir(root, fn); err == nil {
			t.Errorf("WalkDir() did not return an error when called with a non-existent root")
		}
	})
	t.Run("skipall", func(t *testing.T) {
		var ncalls atomic.Int64
		fn := func(path string, d fs.DirEntry, err error) error {
			ncalls.Add(1)
			return fs.SkipAll
		}
		if err := WalkDir(root, fn); err != nil {
			t.Errorf("WalkDir() returned error %v, but expected nil", err)
		}
		if n := ncalls.Load(); n != 1 {
			t.Errorf("WalkDir() called function %v times, but expected one call!", n)
		}
	})
	t.Run("skipone", func(t *testing.T) {
		var ncalls atomic.Int64
		var nskips atomic.Int64
		fn := func(path string, d fs.DirEntry, err error) error {
			t.Logf("fn(): path: %q err: %v", path, err)
			ncalls.Add(1)
			if filepath.Base(path) == "subdir" && d != nil && d.Name() == "subdir" {
				nskips.Add(1)
				return fs.SkipDir
			}
			return nil
		}
		if err := WalkDir(root, fn); err != nil {
			t.Errorf("WalkDir() returned error %v, but expected nil", err)
		}
		if n := nskips.Load(); n != 1 {
			t.Errorf("WalkDir() called function skipped %v dirs, but expected one!", n)
		}
		var expectedCalls int
		for _, input := range inputs {
			if strings.Contains(input.Path, "subdir/") {
				continue
			}
			expectedCalls++
		}
		expectedCalls++ // it'll also be called for the root.
		if n := int(ncalls.Load()); n != expectedCalls {
			t.Errorf("WalkDir() called function %d times, but expected %d calls", n, expectedCalls)
		}
	})
	t.Run("walkDir", func(t *testing.T) {
		var ncalls atomic.Int64
		fn := func(path string, d fs.DirEntry, err error) error {
			ncalls.Add(1)
			if err != nil {
				return fmt.Errorf("called with error: %v", err)
			}
			if path == root {
				if d == nil {
					return fmt.Errorf("called with root, but fs.DirEntry is nil")
				}
				return nil
			}
			// If the path doesn't match, the name building is wrong.
			var testdata TestInput
			for _, input := range inputs {
				if path == input.Path {
					testdata = input
				}
			}
			if testdata.Name == "" {
				return fmt.Errorf("called for path %q not in inputs, or was mangled", path)
			}
			if testdata.DirEntry != nil && d == nil {
				return fmt.Errorf("called for directory %q but called with nil fs.DirEntry", testdata.Name)
			}
			return nil
		}
		if err := WalkDir(root, fn); err != nil {
			t.Errorf("WalkDir() returned error %v, but expected nil", err)
		}
		if n := int(ncalls.Load()); n != len(inputs)+1 {
			t.Errorf("WalkDir() called function %d times, but expected %d calls", n, len(inputs)+1)
		}
	})
}
