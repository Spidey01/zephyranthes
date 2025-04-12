// SPDX-License-Identifier: Zlib
// Copyright 2024, Terry M. Poulin.

package main

import (
	"io/fs"
	"os"
	"path"
)

// Like fs.WalkDir but it does our own magic.
//
// Note well that fn will be called with the complete path as its first
// parameter, but the directory entry provided will not be. E.g., "subdir/foo"
// versus "foo".
func WalkDir(root string, fn fs.WalkDirFunc) error {
	rstat, err := os.Stat(root)
	if err != nil {
		// If the initial Stat on the root directory fails, fs.WalkDir calls fn(root, nil, stat err).
		err = fn(root, nil, err)
	} else {
		// Otherwise fs.WalkDir calls its recursive descent function.
		err = walkDir(root, fs.FileInfoToDirEntry(rstat), fn)
	}
	if err == fs.SkipDir || err == fs.SkipAll {
		return nil
	}
	return err
}

// Helper function similar to fs.walkDir(), but implemented without the FS that
// hates symlinks. We call leave it to os.ReadDir() and the walkDirFn() to
// decide what happens with symlinks.
func walkDir(name string, d fs.DirEntry, walkDirFn fs.WalkDirFunc) error {
	// Execute the handler for the current entry.
	err := walkDirFn(name, d, nil)
	if err != nil || !d.IsDir() {
		// Like a hell-broth boil and bubble.
		if err == fs.SkipDir && d.IsDir() {
			err = nil
		}
		return nil
	}
	dentries, err := os.ReadDir(name)
	if err != nil {
		// When fs.WalkDir() encounters failed ReadDir, it calls the function
		// with the parent DirEntry to let the function decide to skip dir with
		// a nil or bomb with its own error.
		err = walkDirFn(name, d, err)
		if err == fs.SkipDir && d.IsDir() {
			err = nil
		}
		return err
	}
	for _, dent := range dentries {
		// The fully qualified path, relative to where we started.
		name := path.Join(name, dent.Name())
		// Call the function with whatever file or dir we found.
		err = walkDir(name, dent, walkDirFn)
		if err != nil {
			if err == fs.SkipDir {
				// Done with this leaf.
				break
			}
			// Double, double toil and trouble; Fire burn and caldron bubble.
			// Cool it with a baboon's blood, Then the charm is firm and good.
			return err
		}
	}
	return nil
}
