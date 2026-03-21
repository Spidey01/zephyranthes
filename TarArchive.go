// SPDX-License-Identifier: Zlib
// Copyright 2024-2026, Terry M. Poulin.

package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
)

type TarArchive struct {
	file       *os.File
	writer     *tar.Writer
	compressor io.WriteCloser
}

type FilterFunc func(io.Writer) io.WriteCloser

// Creates a new tape archive (tar) at the specified path. If filter is not nil,
// it will be called with the file handle to create a filter. This can be used
// to create a compressed tape archive.
func NewTarArchive(path string, filter FilterFunc) (*TarArchive, error) {
	fp, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	var writer *tar.Writer
	var compressor io.WriteCloser
	if filter != nil {
		compressor = filter(fp)
		writer = tar.NewWriter(compressor)
	} else {
		compressor = nil
		writer = tar.NewWriter(fp)
	}

	return &TarArchive{
		file:       fp,
		writer:     writer,
		compressor: compressor,
	}, nil
}

// Implements [Archive.Name] for tape archives.
func (t *TarArchive) Name() string {
	return t.file.Name()
}

// Implements [Archive.Close] for tape archives.
func (t *TarArchive) Close() error {
	Debugf("Closing Archive %q", t.Name())
	if err := t.writer.Close(); err != nil {
		return err
	}
	if t.compressor != nil {
		if err := t.compressor.Close(); err != nil {
			return err
		}
	}
	if err := t.file.Close(); err != nil {
		return err
	}
	return nil
}

// Implements [Archive.Flush] for tape archives.
func (t *TarArchive) Flush() error {
	Debugf("Flushing Archive %q", t.Name())
	return t.writer.Flush()
}

// Implements [Archive.AddFS] for tape archives.
func (t *TarArchive) AddFS(fsys fs.FS) error {
	return t.writer.AddFS(fsys)
}

// Creates the best possible header from the stat info, and records the file as
// `name` in the header.
func newTarHeader(stat fs.FileInfo, name string) (*tar.Header, error) {
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

// Implements [Archive.AddFile] for tape archives.
func (t *TarArchive) AddFile(fp io.Reader, stat fs.FileInfo, name string) error {
	Debugf("AddFile(): stat.Name(): %q name: %q", stat.Name(), name)
	hdr, err := newTarHeader(stat, name)
	if err != nil {
		return err
	}
	if err = t.writeHeader(hdr); err != nil {
		return nil
	}
	return CopyData(t.writer, FormatName(t, name), fp, name)
}

// Implements [Archive.AddDir] for tape archives.
func (t *TarArchive) AddDir(dp fs.DirEntry, stat fs.FileInfo, name string) error {
	Debugf("AddDirEntry(): stat.Name(): %q name: %q", stat.Name(), name)
	hdr, err := newTarHeader(stat, name)
	if err != nil {
		return err
	}
	if err = t.writeHeader(hdr); err != nil {
		return nil
	}
	return nil
}

// Implements [ExtractableArchive.ExtractTo] for unit testing of tape archives.
func (t *TarArchive) ExtractTo(where string) error {
	// Since we only use this for unit tests, this is just a quick and dirty
	// extraction. Zephyranthes is for backing up files to archives, it doesn't
	// restore them ;).
	fp, err := os.Open(t.Name())
	if err != nil {
		return fmt.Errorf("os.Open(%q) failed: %v", t.Name(), err)
	}
	defer fp.Close()
	// In the same vain, we decide the compressor based on extension. There is
	// no reason to allocate a beefy block of memory for a decompressor to every
	// instance...just to support unit tests.
	var reader *tar.Reader
	if strings.HasSuffix(t.file.Name(), FormatTGZ) || strings.HasSuffix(t.file.Name(), FormatTarGz) {
		filter, err := gzip.NewReader(fp)
		if err != nil {
			return fmt.Errorf("gzip.NewReader failed: %v", err)
		}
		reader = tar.NewReader(filter)
	} else if strings.HasSuffix(t.file.Name(), FormatTZST) || strings.HasSuffix(t.file.Name(), FormatTarZst) {
		filter, err := zstd.NewReader(fp)
		if err != nil {
			return fmt.Errorf("zstd.NewReader failed: %v", err)
		}
		reader = tar.NewReader(filter)
	} else {
		reader = tar.NewReader(fp)
	}
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reader.Next() failed: %v", err)
		}
		path := filepath.Join(where, header.Name)
		// Sanitize the file name to prevent zip slip style overwrites.
		if !strings.HasPrefix(path, filepath.Clean(where)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", path)
		}
		Debugf("Extracting %q -> %q", header.Name, path)
		stat := header.FileInfo()
		if stat.IsDir() {
			if err = os.MkdirAll(path, stat.Mode()); err != nil {
				return fmt.Errorf("extracting directory %q failed: %v", path, err)
			}
			continue
		}
		// Since we're not a general purpose extraction tool, this shouldn't be a problem.
		if err = os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return fmt.Errorf("os.MkdrDir(%q) failed: %v", filepath.Dir(path), err)
		}
		dst, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("os.Create(%q) failed: %v", path, err)
		}
		defer dst.Close()
		if err = CopyData(dst, path, reader, header.Name); err != nil {
			return err
		}
	}
	return nil
}
