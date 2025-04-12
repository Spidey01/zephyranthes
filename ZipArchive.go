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

// Creates a suitable file header based on the stat information. Using
// zip.Writer.Create() on the path instead of providing a real file header,
// default constructs most field, which in turn leads to loss of info like the
// timestamps.
func NewZipHeader(stat fs.FileInfo, name string) (*zip.FileHeader, error) {
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

func (z *ZipArchive) AddFile(fp *os.File, stat os.FileInfo, name string) error {
	Debugf("AddFile(): fp.Name(): %q stat.Size(): %d name: %q", fp.Name(), stat.Size(), name)
	hdr, err := NewZipHeader(stat, name)
	if err != nil {
		return err
	}
	w, err := z.writer.CreateHeader(hdr)
	if err != nil {
		return err
	}
	return CopyData(w, FormatName(z, name), fp, fp.Name())
}

func (z *ZipArchive) AddDir(dp fs.DirEntry, stat os.FileInfo, name string) error {
	Debugf("AddDirEntry(): dp.Name(): %q stat.Size(): %d name: %q", dp.Name(), stat.Size(), name)
	path := name
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	hdr, err := NewZipHeader(stat, path)
	if err != nil {
		return err
	}
	// We ignore the returned writer as there are no file contents for a directory.
	_, err = z.writer.CreateHeader(hdr)
	if err != nil {
		return err
	}
	return nil
}
