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
	// Takes care of creating a suitable file header, but we may want to get fancier.
	w, err := z.writer.Create(out)
	if err != nil {
		return err
	}
	return CopyData(w, FormatName(z, out), fp, fp.Name())
}

func (z *ZipArchive) AddDir(in, out string) error {
	Verbosef("AddDir(%q, %q)", in, out)
	return WalkDir(in, z.walkDirFunc)
}

func (z *ZipArchive) walkDirFunc(path string, d fs.DirEntry, err error) error {
	Verbosef("walkDirFunc(%s, %s, %v)", path, fs.FormatDirEntry(d), err)
	if err != nil {
		Verbosef("walkDirFunc called /w err=%v", err)
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
