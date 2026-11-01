// SPDX-License-Identifier: Zlib
// Copyright 2024-2026, Terry M. Poulin.
package main

import (
	"context"
	_ "embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"runtime/debug"
	"slices"
	"strings"
)

var options = NewOptions()

//go:embed zephyr.1.md
var manual string

//go:embed LICENSE.txt
var license string

//go:embed 3rd_party_licenses.txt
var third_party_licenses string

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
			err = runBackupSpec(ctx, spec)
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

// Handles one backup spec. Ensures the before command hook runs, and provided
// that suceeds, calls backup() and then always runs the after command hook.
func runBackupSpec(ctx context.Context, spec BackupSpec) error {
	Verbosef("Running before commands...")
	if err := runCommands(ctx, spec.Before); err != nil {
		return err
	}
	backupErr := backup(ctx, spec)
	Verbosef("Running after commands...")
	if err := runCommands(ctx, spec.After); err != nil {
		return err
	}
	return backupErr
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

// commands returns the set of exec.Cmds for the provided spec.
func commands(ctx context.Context, spec CommandSpec) []*exec.Cmd {
	// [os.Getenv] doesn't have a varient overriding [os.Environ]. To call
	// [os.Expand] we need to provide our own that first checks the overrides in
	// [CommandSpec.Environment] or we need to build/restore our own env.
	getenv := func(key string) string {
		if value, exists := spec.Environment[key]; exists {
			return value
		}
		return os.Getenv(key)
	}

	// A non-nil [exec.Cmd.Env] replaces [os.Environ], to override we need to
	// merge [CommandSpec.Environment] and compute the entire environment.
	environ := slices.Clone(os.Environ())
	for key, value := range spec.Environment {
		environ = append(environ, key+"="+value)
	}

	// Parse each command as a word separate list subject to variable expansion.
	var result []*exec.Cmd
	for _, line := range spec.Commands {
		var name string
		var args []string
		for _, elem := range strings.Split(line, " ") {
			switch {
			case name == "":
				name = elem
			default:
				if a := os.Expand(elem, getenv); a != "" {
					args = append(args, a)
				}
			}
		}
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Env = environ
		cmd.Dir = spec.Directory
		result = append(result, cmd)
	}
	return result
}

// runCommands executes the provided command spec. Returns an error on failure.
// Halts at the first failure. Options can override the execution behavior.
//
// - options.NoExec: commands are neither printed nor executed.
// - options.DryRun: commands are printed, not executed.
// - options.Verbose: commands are executed, output printed.
//
// If spec.Timeout is zero then no timeout is applied.
func runCommands(parent context.Context, spec CommandSpec) error {
	ctx := parent
	var cancel context.CancelFunc
	if spec.Timeout > 0 {
		ctx, cancel = context.WithTimeout(parent, spec.Timeout)
		defer cancel()
	}

	for lineno, cmd := range commands(ctx, spec) {
		Debugf("runCommand: parsed spec.Commands[%d]: %q", lineno, cmd)
		if options.NoExec {
			continue
		}
		line := strings.Join(cmd.Args, " ")
		Infof("Running command %s", line)
		if options.DryRun {
			fmt.Println(line)
			continue
		}

		if options.Verbose {
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
		}

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command %q failed: %w", line, err)
		}
	}
	return nil
}
