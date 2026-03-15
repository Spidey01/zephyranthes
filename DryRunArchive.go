// SPDX-License-Identifier: Zlib
// Copyright 2026, Terry M. Poulin.

package main

import (
	"fmt"
	"io"
	"io/fs"
	"strings"
)

// Creates a phony archive that can be used for `--dry-run` mode and for unit
// testing that just needs to access stored paths or force an error return.
type DryRunArchive struct {
	name     string
	err      error
	contents []string
}

func NewDryRunArchive(name string) *DryRunArchive {
	Debugf("Opening Archive %q", name)
	return &DryRunArchive{
		name: name,
	}
}

func (dry *DryRunArchive) Name() string {
	return dry.name
}

func (dry *DryRunArchive) Close() error {
	Debugf("Closing Archive %q", dry.name)
	return dry.err
}

func (dry *DryRunArchive) Flush() error {
	Debugf("Flushing Archive %q", dry.name)
	return dry.err
}

func (dry *DryRunArchive) AddFS(fsys fs.FS) error {
	if dry.err != nil {
		return dry.err
	}
	// AddFS is not used anywhere, yet.
	return fmt.Errorf("not implemented: DryRunArchive.AddFS()")
}

func (dry *DryRunArchive) AddFile(fp io.Reader, stat fs.FileInfo, name string) error {
	Debugf("AddFile(): stat.Name(): %q name: %q", stat.Name(), name)
	if dry.err != nil {
		return dry.err
	}
	if stat.IsDir() {
		// real implementations error here due to the archive format.
		return fmt.Errorf("cannot add directory %q as file", name)
	}
	dry.contents = append(dry.contents, name)
	return nil
}

func (dry *DryRunArchive) AddDir(dp fs.DirEntry, stat fs.FileInfo, name string) error {
	Debugf("AddDirEntry(): stat.Name(): %q name: %q", stat.Name(), name)
	if dry.err != nil {
		return dry.err
	}
	if !stat.IsDir() {
		// real implementations error here due to the archive format.
		return fmt.Errorf("cannot add file %q as directory", name)
	}
	path := name
	// ZipArchive uses this convention, we use it here because it makes it
	// obvious without having to save the fs.FileInfo.
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	dry.contents = append(dry.contents, path)
	return nil
}
