// SPDX-License-Identifier: Zlib
// Copyright 2026, Terry M. Poulin.
package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// Bootstrap options and logging. Returns the file handle for the log file.
func setupTestOptionsAndLogging(t *testing.T) *os.File {
	lfp, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("os.CreateTemp() failed: %v", err)
	}
	t.Logf("See %q for log file output", lfp.Name())

	args := []string{"--dry-run", "log-level", "debug", "-log-file", lfp.Name()}
	if err := options.Parse(args); err != nil {
		t.Fatalf("Failed parsing options: %v", err)
	}
	if err := SetupLogging(t.Context(), options.Name(), options.LogLevel, options.LogFile); err != nil {
		t.Fatalf("SetupLogging() failed: %v", err)
	}

	return lfp
}

func TestBackup(t *testing.T) {
	lfp := setupTestOptionsAndLogging(t)
	defer lfp.Close()
	// N.B. we don't remove the file via defer, we only want to remove if test successful.

	// Populate test data.
	testdata, cleanup := MakeTestDir(t)
	defer cleanup()
	inputs := MakeTestData(t, testdata)

	// Build a backup spec from the input.
	spec := BackupSpec{
		Name: t.Name(),
		// Since --dry-run is used, path and format are just needed for validation.
		Path:   filepath.Join(testdata, "archive.dryrun"),
		Format: FormatZip,
	}
	for _, input := range inputs {
		spec.Contents = append(spec.Contents, input.Path)
	}

	// First do the whole backup. We can't verify the contents this way since
	// there is no access to the DryRunArchive, but it's a good test run.
	//
	// We could use a real implementation like ZipArchive and verify it, but
	// that's largely covered by TestArchive, and the operations in-between have
	// their own tests. So, this is good enough!
	Debugf("Running backup() from %q", t.Name())
	if err := backup(t.Context(), spec); err != nil {
		t.Errorf("backup() failed: %v", err)
	}
	Debugf("---- end of backup() ---- ")

	os.Remove(lfp.Name()) // only clean up log on success.
}

func TestBackupFile(t *testing.T) {
	lfp := setupTestOptionsAndLogging(t)
	defer lfp.Close()
	// N.B. we don't remove the file via defer, we only want to remove if test successful.

	// Populate test data.
	testdata, cleanup := MakeTestDir(t)
	defer cleanup()
	inputs := MakeTestData(t, testdata)

	archive := NewDryRunArchive(filepath.Join(testdata, "archive.dryrun"))

	// First, let's test the error cases
	if err := backupFile(archive, nil, "/does/not/exist"); err == nil {
		t.Errorf("backupFile() with nil file info and non-existent file did not give an error")
	}

	// Let's hit each test input, noting success for files and failures for non-files.
	for _, input := range inputs {
		err := backupFile(archive, input.Stat, input.Path)
		switch {
		case err == nil && input.DirEntry != nil:
			t.Errorf("backupFile(..., %q) with directory did not give an error", input.Path)
		case err != nil && input.DirEntry == nil:
			t.Errorf("backupFile(..., %q) failed: %v", input.Path, err)
		case err == nil && len(archive.contents) > 1:
			file := archive.contents[len(archive.contents)-1]
			// This will be equal to path, because that's what we used above; not the relative name.
			if file != input.Path {
				t.Errorf("backupFile(..., %q) added file as %q rather than %q", input.Path, file, input.Path)
			} else {
				t.Logf("backupFile(..., %q) added file as %q", input.Path, file)
			}
		}
	}

	os.Remove(lfp.Name()) // only clean up log on success.
}

func TestBackupDir(t *testing.T) {
	lfp := setupTestOptionsAndLogging(t)
	defer lfp.Close()
	// N.B. we don't remove the file via defer, we only want to remove if test successful.

	// Populate test data.
	testdata, cleanup := MakeTestDir(t)
	defer cleanup()
	inputs := MakeTestData(t, testdata)

	archive := NewDryRunArchive(filepath.Join(testdata, "archive.dryrun"))

	// First, let's test the error cases
	if err := backupDir(archive, "/does/not/exist"); err == nil {
		t.Errorf("backupDir() with non-existent directory did not give an error")
	}

	// Let's hit each test input. Technically, backupDir doesn't care if its
	// called for a file because of implementation details. So, we can't test
	// that backupDir() on a file is an error--because that's not how it
	// actually works! It recurses adding the tree.
	for _, input := range inputs {
		if strings.Contains(input.Name, "/") {
			// skip subdirs because of the recursion.
			continue
		}
		if err := backupDir(archive, input.Path); err != nil {
			t.Errorf("backupDir(..., %q) failed: %v", input.Path, err)
		}
	}

	var nFilesFound int
	var nDirsFound int
	for _, path := range archive.contents {
		// symlinks are stored as files, but the TestInput has to record it as a
		// directory for other test purposes. Thus, it becomes a special case
		// here.
		if strings.HasSuffix(path, "/") {
			nDirsFound++
			t.Logf("actual input dir: %v", path)
		} else {
			t.Logf("actual input file: %v", path)
			nFilesFound++
		}
	}
	var nExpectedFiles int
	var nExpectedDirs int
	for _, input := range inputs {
		// We also have to consider dirlinkup in the fixture. For the purposes
		// of backupDir this is a file (symlink), but the TestInput has to
		// record it as a DirEntry for other purposes. Since this is the only
		// case that cares that directory symlinks are stored as files, we
		// special case it in this test. It's not worth handling more generally
		// when it's the only one.
		if strings.HasSuffix(input.Name, "dirlinkup") {
			t.Logf("expected input link: %v", input.Name)
			nExpectedFiles++
		} else if input.DirEntry == nil {
			t.Logf("expected input file: %v", input.Name)
			nExpectedFiles++
		} else {
			t.Logf("expected input dir: %v", input.Name)
			nExpectedDirs++
		}
	}
	if nFilesFound != nExpectedFiles {
		t.Errorf("Expected %d files archived, but %d files archived", nFilesFound, nExpectedFiles)
	}
	if nDirsFound != nExpectedDirs {
		t.Errorf("Expected %d directories archived, but %d directories archived", nDirsFound, nExpectedDirs)
	}

	os.Remove(lfp.Name()) // only clean up log on success.
}

func TestCommands(t *testing.T) {
	spec := CommandSpec{
		Commands: []string{
			"multi word command",
			"another command",
			"expanded $foo command",
			"with   spaces",
			"not $recursive",
		},
		Directory: "/override/working/directory",
		Environment: map[string]string{
			"foo":       "bar",
			"recursive": "$foo",
			"SHELL":     "overriden",
		},
		Timeout: 1 * time.Second,
	}
	var count int
	for i, cmd := range commands(t.Context(), spec) {
		count++
		if cmd.Dir != spec.Directory {
			t.Errorf("Index %d: Expected Cmd.Dir %s, got %s", i, cmd.Dir, spec.Directory)
		}
		assert := func(expected, actual []string) {
			t.Helper()
			if !slices.Equal(actual, expected) {
				t.Errorf("Index %d: Expected args %+v, but parsed %+v", i, expected, actual)
			}
		}
		switch i {
		case 0:
			assert(cmd.Args, []string{"multi", "word", "command"})
		case 1:
			assert(cmd.Args, []string{"another", "command"})
		case 2:
			assert(cmd.Args, []string{"expanded", "bar", "command"})
		case 3:
			for i, a := range cmd.Args {
				t.Logf("cmd.Args[%d]: %q", i, a)
			}
			assert(cmd.Args, []string{"with", "spaces"})
		case 4:
			assert(cmd.Args, []string{"not", "$foo"})
		default:
			t.Fatalf("Test doesn't have index %d", i)
		}
		for key, value := range spec.Environment {
			env := key + "=" + value
			if !slices.Contains(cmd.Env, env) {
				t.Errorf("Index %d: Expected %s in Cmd.Env%+v", i, env, cmd.Env)
			}
			if !slices.Contains(cmd.Environ(), env) {
				t.Errorf("Index %d: Expected %s in Cmd.Environ()%+v", i, env, cmd.Env)
			}
		}
	}
	if want := len(spec.Commands); count != want {
		t.Errorf("Expected %d exec.Cmd objects, but %d created", want, count)
	}
}
