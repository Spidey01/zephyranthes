// SPDX-License-Identifier: Zlib
// Copyright 2026, Terry M. Poulin.

package main

import (
	"compress/gzip"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
)

// Since we can't pass a function like gzip.NewWriter to TarArchive directly due
// to how Go treats function pointers, wrapper functions are required. In some
// formats, this is more necessary than others. Also, we need readers to support
// the unit tests.

func newGzipWriter(w io.Writer) (io.WriteCloser, error) {
	return gzip.NewWriter(w), nil
}

func newGzipReader(r io.Reader) (*gzip.Reader, error) {
	filter, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gzip.NewReader failed: %v", err)
	}
	return filter, err
}

func newZstdWriter(w io.Writer) (io.WriteCloser, error) {
	encoder, err := zstd.NewWriter(w)
	if err != nil {
		return nil, fmt.Errorf("zstd.NewWriter failed: %v", err)
	}
	return encoder, nil
}

func newZstdReader(r io.Reader) (*zstd.Decoder, error) {
	filter, err := zstd.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gzip.NewReader failed: %v", err)
	}
	return filter, err
}
