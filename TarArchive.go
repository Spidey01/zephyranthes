// SPDX-License-Identifier: Zlib
// Copyright 2024, Terry M. Poulin.

package main

import (
	"archive/tar"
	"io/fs"
	"os"
	"strings"
)

type TarArchive struct {
	Archive
	file   *os.File
	writer *tar.Writer
}

func NewTarArchive(path string) (*TarArchive, error) {
	fp, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return &TarArchive{
		file:   fp,
		writer: tar.NewWriter(fp),
	}, nil
}

func (t *TarArchive) Name() string {
	return t.file.Name()
}

func (t *TarArchive) Close() error {
	if err := t.writer.Close(); err != nil {
		return err
	}
	if err := t.file.Close(); err != nil {
		return err
	}
	return nil
}

func (t *TarArchive) Flush() error {
	return t.writer.Flush()
}

func (t *TarArchive) AddFS(fsys fs.FS) error {
	return t.writer.AddFS(fsys)
}

func (t *TarArchive) AddFile(in, out string) error {
	fp, err := os.Open(in)
	if err != nil {
		return err
	}
	defer fp.Close()
	stat, err := fp.Stat()
	if err != nil {
		return err
	}
	hdr, err := NewTarHeader(in, stat, out)
	if err != nil {
		return err
	}
	if err = t.writer.WriteHeader(hdr); err != nil {
		return err
	}
	return CopyData(t.writer, FormatName(t, out), fp, in)
}

func (t *TarArchive) walkDirFunc(path string, d fs.DirEntry, err error) error {
	Verbosef("walkDirFunc(%s, %s, %v)", path, fs.FormatDirEntry(d), err)
	if err != nil {
		Verbosef("walkDirFunc called /w err=%v", err)
	}
	st, err := d.Info()
	if err != nil {
		Warningf("Skipping %v", path)
		if d.IsDir() {
			return fs.SkipDir
		}
		return nil
	}
	hdr, err := NewTarHeader(path, st, path)
	if err != nil {
		return err
	}
	Verbosef("+ %s (%s)", hdr.Name, hdr.Linkname)
	if err = t.writer.WriteHeader(hdr); err != nil {
		return err
	}
	if !d.IsDir() {
		// return t.AddFile(path, path)
		fp, err := os.Open(path)
		if err != nil {
			return err
		}
		return CopyData(t.writer, FormatName(t, path), fp, path)
	}
	return nil
}

func (t *TarArchive) AddDir(in, out string) error {
	Verbosef("AddDir(%q, %q)", in, out)
	return WalkDir(in, t.walkDirFunc)
}

// Creates the best possible header from the stat info, and records the file as
// `name` in the header.
func NewTarHeader(src string, stat fs.FileInfo, name string) (*tar.Header, error) {
	var err error

	// Since stat may not give us a suitable name we can open (foo rather than
	// subdir/foo), we need to do our own determination on the link name from
	// the real source name.
	var linkName string
	if stat.Mode().Type()&fs.ModeSymlink != 0 {
		linkName, err = os.Readlink(src)
		if err != nil {
			Warningf("ReadLink: %v", err)
		}
	}

	// This takes care of setting the obvious fields from stat and linkName:
	// - Name, ModTimes, Mod
	// Plus based on stat.Mode():
	// - Typeflag, Size, Linkname
	// And where available from stat.Sys():
	// - Uid, Guid, Uname, Gname, AccessTime, ChangeTime, Xattrs, PAXRecords.
	//
	// Go's implementation always sets the obvious stat fields before an error occurs. The only time an
	// error should be expected is if the file type is unsupported or an error
	// occurs looking up ownership.
	hdr, err := tar.FileInfoHeader(stat, linkName)
	if err == nil {
		hdr.Name = name
		if stat.IsDir() && !strings.HasSuffix(hdr.Name, "/") {
			hdr.Name += "/"
		}
	}
	return hdr, err
}
