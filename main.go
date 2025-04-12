// SPDX-License-Identifier: Zlib
// Copyright 2024-2025, Terry M. Poulin.
package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
)

var options = NewOptions()

func main() {
	options.MustParseArgs()
	SetupLogging(options.Name(), options.LogLevel, options.LogFile)
	for _, arg := range options.Args() {
		Verbosef("Parsing %s", arg)
		specs, err := BackupSpecsFromFile(arg)
		if err != nil {
			Die("unable to load %s\n%v\n", arg, err)
		}
		for i, spec := range specs {
			Verbosef("Running backup %d: %s", i, spec)
			err = backup(context.Background(), spec)
			if err != nil {
				Fatalf("Backup %s failed: %v", spec.Name, err)
			}
		}
	}
}

// Executes the backup specification using the provided context. Returns nil
// once the job is complete, or an error is the operation failed.
func backup(ctx context.Context, spec BackupSpec) error {
	archive, err := CreateArchive(spec.Path, spec.Format)
	if err != nil {
		return err
	}
	defer archive.Close()
	Verbosef("Archiving contents...")
	for _, fn := range spec.Contents {
		if err = ctx.Err(); err != nil {
			return err
		}
		stat, err := os.Stat(fn)
		if err != nil {
			Warningf("Skipping %s: %v", fn, err)
			continue
		}
		Verbosef("Inspecting %s", fs.FormatFileInfo(stat))
		if stat.IsDir() {
			Infof("Adding directory tree %s", stat.Name())
			err = backupDir(archive, fs.FileInfoToDirEntry(stat))
		} else {
			Infof("Adding file %s", stat.Name())
			err = backupFile(archive, stat, fn)
		}
		if err != nil {
			Errorf("Failed to backup %s: %v", fn, err)
			break
		}
	}
	return nil
}

// Adds the specified file to the archive.
func backupFile(archive Archive, stat fs.FileInfo, path string) error {
	fp, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fp.Close()
	if options.DryRun {
		return nil
	}
	return archive.AddFile(fp, stat, path)
}

// Recursively adds the specified root to the archive.
func backupDir(archive Archive, root fs.DirEntry) error {
	fn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// N.B. if err is set, d is nil.
			Debugf("walkDirFunc(%s, nil, %v)", path, err)
			// We can return nil or fs.SkipDir/fs.SkipAll to ignore this tree,
			// or an error to bork the operation.
			return err
		} else {
			Debugf("walkDirFunc(%s, %s, %v)", path, fs.FormatDirEntry(d), err)
		}
		// Since it's valid on files and directories, we can stat before caring
		// which it references.
		stat, err := d.Info()
		if err != nil {
			return fmt.Errorf("stat %q failed: %w", path, err)
		}
		if !d.IsDir() {
			return backupFile(archive, stat, path)
		}
		if options.DryRun {
			return nil
		}
		return archive.AddDir(d, stat, path)
	}
	return WalkDir(root.Name(), fn)
}
