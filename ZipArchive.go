// SPDX-License-Identifier: Zlib
// Copyright 2024, Terry M. Poulin.

package main

import (
	"archive/zip"
	"io/fs"
	"os"
	"strings"
)

type ZipArchive struct {
	Archive
	file   *os.File
	writer *zip.Writer
}

// Creates a new zip archive at the specified path.
func NewZipArchive(path string) (*ZipArchive, error) {
	fp, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return &ZipArchive{
		file:   fp,
		writer: zip.NewWriter(fp),
	}, nil
}

func (z *ZipArchive) Name() string {
	return z.file.Name()
}

func (z *ZipArchive) Close() error {
	if err := z.writer.Close(); err != nil {
		return err
	}
	if err := z.file.Close(); err != nil {
		return err
	}
	return nil
}

func (z *ZipArchive) Flush() error {
	return z.writer.Flush()
}

func (z *ZipArchive) AddFS(fsys fs.FS) error {
	return z.writer.AddFS(fsys)
}

func (z *ZipArchive) AddFile(in, out string) error {
	fp, err := os.Open(in)
	if err != nil {
		return err
	}
	defer fp.Close()
	stat, err := fp.Stat()
	if err != nil {
		return err
	}
	hdr, err := NewZipHeader(in, stat, out)
	if err != nil {
		return err
	}
	w, err := z.writer.CreateHeader(hdr)
	if err != nil {
		return err
	}
	return CopyData(w, FormatName(z, out), fp, fp.Name())
}

func (z *ZipArchive) AddDir(in, out string) error {
	Debugf("AddDir(%q, %q)", in, out)
	return WalkDir(in, z.walkDirFunc)
}

// Callback for our WalkDir function. Used to add discovered directories and
// files to the archive.
func (z *ZipArchive) walkDirFunc(path string, d fs.DirEntry, err error) error {
	Debugf("walkDirFunc(%s, %s, %v)", path, fs.FormatDirEntry(d), err)
	if err != nil {
		Debugf("walkDirFunc called /w err=%v", err)
	}
	if !d.IsDir() {
		return z.AddFile(path, path)
	} else {
		name := path
		if !strings.HasSuffix(name, "/") {
			name += "/"
		}
		_, err := z.writer.Create(name)
		if err != nil {
			return err
		}
	}
	return nil
}

// Creates a suitable file header based on the stat information. Using
// zip.Writer.Create() on the path instead of providing a real file header,
// default constructs most field, which in turn leads to loss of info like the
// timestamps.
func NewZipHeader(src string, stat fs.FileInfo, name string) (*zip.FileHeader, error) {
	// This handles setting the fields related to uncompressed size and timestamps.
	hdr, err := zip.FileInfoHeader(stat)
	if err != nil {
		return nil, err
	}
	// Ensure the name is built correctly. E.g., subdir/foo rather than foo.
	hdr.Name = name
	hdr.Method = zip.Deflate
	return hdr, nil
}
