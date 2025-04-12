// SPDX-License-Identifier: Zlib
// Copyright 2024, Terry M. Poulin.

package main

import (
	"archive/tar"
	"io"
	"io/fs"
	"os"
	"strings"
)

type TarArchive struct {
	Archive
	file   *os.File
	writer *tar.Writer
}

// Creates a new tape archive (tar) at the specified path.
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

// Creates the best possible header from the stat info, and records the file as
// `name` in the header.
func NewTarHeader(stat fs.FileInfo, name string) (*tar.Header, error) {
	var err error

	// Since the Name() method on file/direntry/fileinfo structures typically
	// return the base name (foo) rather than the real path (subdir/foo), we use
	// that for attempting to read link info.

	var linkName string
	if stat.Mode().Type()&fs.ModeSymlink != 0 {
		linkName, err = os.Readlink(name)
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

func (t *TarArchive) writeHeader(hdr *tar.Header) error {
	Verbosef("+ %s (%s)", hdr.Name, hdr.Linkname)
	if err := t.writer.WriteHeader(hdr); err != nil {
		return err
	}
	return nil
}

func (t *TarArchive) AddFile(fp io.Reader, stat fs.FileInfo, name string) error {
	Debugf("AddFile(): stat.Name(): %q name: %q", stat.Name(), name)
	hdr, err := NewTarHeader(stat, name)
	if err != nil {
		return err
	}
	if err = t.writeHeader(hdr); err != nil {
		return nil
	}
	return CopyData(t.writer, FormatName(t, name), fp, name)
}

func (t *TarArchive) AddDir(dp fs.DirEntry, stat fs.FileInfo, name string) error {
	Debugf("AddDirEntry(): stat.Name(): %q name: %q", stat.Name(), name)
	hdr, err := NewTarHeader(stat, name)
	if err != nil {
		return err
	}
	if err = t.writeHeader(hdr); err != nil {
		return nil
	}
	return nil
}
