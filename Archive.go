// SPDX-License-Identifier: Zlib
// Copyright 2024, Terry M. Poulin.

package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
)

type Archive interface {
	// Returns the path name of the archive.
	Name() string
	// Finishes the archive.
	Close() error
	// Flush the current state of data.
	Flush() error
	// AddFS adds the files from fs.FS to the archive. It walks the directory
	// tree starting at the root of the filesystem adding each file to the
	// root of the archive while maintaining the directory structure.
	AddFS(fsys fs.FS) error
	// Creates a file entry. Data is copied from `fp` into the archive as `name`
	// based on the original information provided in `stat`. Name should refer
	// to the desired path in the archive (e.g., subdir/foo) and for openning
	// from the current directory. It is expected that any parent directories
	// relevant to `name` have already been created with AddDir(). The caller is
	// responsible for closing fp.
	AddFile(fp *os.File, stat os.FileInfo, name string) error
	// Creates a directory entry. A directory record is added to the archive as
	// `name` using the original information provided by `stat`. This is a non
	// recursive operation, and it is expected that any parent directory already
	// has an entry.
	AddDir(dp fs.DirEntry, stat os.FileInfo, name string) error
}

// Factory function returning the correct Archive implementation for format.
func CreateArchive(name, format string) (Archive, error) {
	switch format {
	case FormatTar:
		return NewTarArchive(name)
	case FormatZip:
		return NewZipArchive(name)
	default:
		return nil, fmt.Errorf("unsupported backup format: %s", format)
	}
}

// Returns a string formatted in the form "archive name:file name" suitable to
// reference some content `name` within the provided `archive` handle.
func FormatName(archive Archive, name string) string {
	return fmt.Sprintf("%s:%s", archive.Name(), name)
}

// Performs an io.Copy() from dst to src. If an error occurs, it will be wrapped
// in an error described by dname and sname as the destination and source name
// respectively.
func CopyData(dst io.Writer, dname string, src io.Reader, sname string) error {
	nb, err := io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("error: %v source: %q destination: %q bytes copied: %d",
			err, sname, dname, nb)
	}
	return nil
}
