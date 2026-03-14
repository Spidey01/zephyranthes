// SPDX-License-Identifier: Zlib
// Copyright 2024-2026, Terry M. Poulin.

package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type ZipArchive struct {
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

// Implements [Archive.Name] for ZIP archives.
func (z *ZipArchive) Name() string {
	return z.file.Name()
}

// Implements [Archive.Close] for ZIP archives.
func (z *ZipArchive) Close() error {
	Debugf("Closing Archive %q", z.Name())
	if err := z.writer.Close(); err != nil {
		return err
	}
	if err := z.file.Close(); err != nil {
		return err
	}
	return nil
}

// Implements [Archive.Flush] for ZIP archives.
func (z *ZipArchive) Flush() error {
	Debugf("Flushing Archive %q", z.Name())
	return z.writer.Flush()
}

// Implements [Archive.AddFS] for ZIP archives.
func (z *ZipArchive) AddFS(fsys fs.FS) error {
	return z.writer.AddFS(fsys)
}

// Creates a suitable file header based on the stat information. Using
// zip.Writer.Create() on the path instead of providing a real file header,
// default constructs most field, which in turn leads to loss of info like the
// timestamps.
func newZipHeader(stat fs.FileInfo, name string) (*zip.FileHeader, error) {
	// Info-Zip and a few others have a means of storing Unix symbolic links in
	// the archive, but I'm not familiar with this extension, and Go's
	// implementation doesn't seem to support it.
	if stat.Mode().Type()&fs.ModeSymlink != 0 {
		linkDestination, err := os.Readlink(name)
		if err != nil {
			Errorf("reading symlink %q failed: %v", name, err)
			linkDestination = ""
		}
		Warningf("Archive member %q refers to a symlink to %q and will be stored as that file's contents rather than as a symbolic link.",
			name, linkDestination)
		Verbosef("To work around this ")
	}
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

// Implements [Archive.AddFile] for ZIP archives.
func (z *ZipArchive) AddFile(fp io.Reader, stat fs.FileInfo, name string) error {
	Debugf("AddFile(): stat.Name(): %q name: %q", stat.Name(), name)
	hdr, err := newZipHeader(stat, name)
	if err != nil {
		return err
	}
	w, err := z.writer.CreateHeader(hdr)
	if err != nil {
		return err
	}
	return CopyData(w, FormatName(z, name), fp, name)
}

// Implements [Archive.AddDir] for ZIP archives.
func (z *ZipArchive) AddDir(dp fs.DirEntry, stat fs.FileInfo, name string) error {
	Debugf("AddDirEntry(): stat.Name(): %q name: %q", stat.Name(), name)
	path := name
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	hdr, err := newZipHeader(stat, path)
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

// Implements [ExtractableArchive.ExtractTo] for unit testing of ZIP archives.
func (z *ZipArchive) ExtractTo(where string) error {
	// Since this is only for unit tests, just make a separate reader instead of
	// seeking back.
	reader, err := zip.OpenReader(z.Name())
	if err != nil {
		return fmt.Errorf("zip.OpenReader(%q) failed: %v", z.Name(), err)
	}
	defer reader.Close()
	for _, file := range reader.File {
		path := filepath.Join(where, file.Name)
		// Sanitize the file name to prevent zip slip style overwrites.
		if !strings.HasPrefix(path, filepath.Clean(where)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", path)
		}
		Debugf("Extracting %q -> %q", file.Name, path)

		if file.FileInfo().IsDir() {
			if err = os.MkdirAll(path, file.Mode()); err != nil {
				return fmt.Errorf("extracting directory %q failed: %v", path, err)
			}
			continue
		}

		// This might get perms wrong if the slice isn't ordered to have parent
		// directories first, but we're extending the codec to support unit
		// testing not writing a real archive extractor.
		if err = os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return fmt.Errorf("os.MkdrDir(%q) failed: %v", filepath.Dir(path), err)
		}
		src, err := file.Open()
		if err != nil {
			return fmt.Errorf("file.Open() for %q failed: %v", file.Name, err)
		}
		// For the same reason, we don't expect test data to backup pending file
		// closes that drastically :P.
		defer src.Close()
		dst, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("os.Create(%q) failed: %v", path, err)
		}
		defer dst.Close()
		if err = CopyData(dst, path, src, file.Name); err != nil {
			return err
		}
	}
	return nil
}
