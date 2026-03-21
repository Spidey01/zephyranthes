// SPDX-License-Identifier: Zlib
// Copyright 2024-2026, Terry M. Poulin.
package main

import (
	"context"
	_ "embed"
	"fmt"
	"io/fs"
	"os"
	"runtime/debug"
)

var options = NewOptions()

//go:embed zephyr.1.md
var manual string

//go:embed LICENSE.txt
var license string

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	options.MustParseArgs()
	if printInfoFlags() {
		os.Exit(0)
	}
	SetupLogging(ctx, options.Name(), options.LogLevel, options.LogFile)
	if options.Directory != "" {
		if err := os.Chdir(options.Directory); err != nil {
			Fatalf("Changing directory to %q failed: %v", options.Directory, err)
		}
		Verbosef("Changed directory to %q", options.Directory)
	}
	for _, arg := range options.Args() {
		Verbosef("Parsing %s", arg)
		specs, err := BackupSpecsFromFile(arg)
		if err != nil {
			Die("unable to load %s\n%v\n", arg, err)
		}
		for i, spec := range specs {
			Verbosef("Running backup %d: %s", i, spec)
			err = backup(ctx, spec)
			if err != nil {
				Fatalf("Backup %q failed: %v", spec.Name, err)
			}
		}
	}
}

// Handles various flags that result in a print and exit operation.
// - version info
// - man page
// - copyright and license info.
func printInfoFlags() bool {
	switch {
	case options.Version:
		printVersionInfo()
		return true
	case options.ManPage:
		fmt.Println(manual)
		return true
	case options.About:
		fmt.Println(license)
		fmt.Println(third_party_licenses)
		return true
	default:
		return false
	}
}

func printVersionInfo() {
	version := "unknown"
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		version = buildInfo.Main.Version
	}
	fmt.Println("Zephyranthes version", version)
}

// Executes the backup specification using the provided context. Returns nil
// once the job is complete, or an error is the operation failed.
func backup(ctx context.Context, spec BackupSpec) error {
	archive, err := CreateArchive(os.ExpandEnv(spec.Path), spec.Format)
	if err != nil {
		return err
	}
	defer archive.Close()
	Verbosef("Archiving contents...")
	for _, fn := range spec.Contents {
		if err = ctx.Err(); err != nil {
			return err
		}
		fn = os.ExpandEnv(fn)
		stat, err := os.Stat(fn)
		if err != nil {
			Warningf("Skipping %s: %v", fn, err)
			continue
		}
		Verbosef("Inspecting %s (%s)", fn, fs.FormatFileInfo(stat))
		if stat.IsDir() {
			Infof("Adding directory tree %s", fn)
			err = backupDir(archive, fn)
		} else {
			Infof("Adding file %s", fn)
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
	return archive.AddFile(fp, stat, path)
}

// Recursively adds the specified root to the archive.
func backupDir(archive Archive, root string) error {
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
		// N.B. this includes symlinks regardless of target.
		if !d.IsDir() {
			return backupFile(archive, stat, path)
		}
		return archive.AddDir(d, stat, path)
	}
	return WalkDir(root, fn)
}
