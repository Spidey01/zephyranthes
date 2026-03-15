// SPDX-License-Identifier: Zlib
// Copyright 2026, Terry M. Poulin.

package main

import "testing"

func assertOptionsParse(t *testing.T, opts *Options, args []string) bool {
	if err := opts.Parse(args); err != nil {
		t.Errorf("Options.Parse(%+v) failed: %v", args, err)
		return true
	}
	if opts.Args() == nil {
		t.Errorf("Options.Args() is nil after parsing empty args")
		return true
	}
	return false
}

func assertOptionsFlag[Flag bool | string | ~int](t *testing.T, actual *Flag, expected Flag, args []string) bool {
	if *actual != expected {
		t.Errorf("Options.Parse(%+v): actual: %v, but expected: %v", args, *actual, expected)
		return true
	}
	return false
}

func TestOptions(t *testing.T) {
	opts := NewOptions()
	if opts.FlagSet == nil {
		t.Errorf("NewOptions")
	}
	if opts.FlagSet.Usage == nil {
		t.Errorf("No usage function set")
	}
	t.Run("help", func(t *testing.T) {
		help := map[string]bool{"-h": true, "-help": true, "--help": true, "": false, "-v": false}
		for arg, expected := range help {
			opts.Help = false
			args := []string{arg}
			assertOptionsParse(t, opts, args)
			assertOptionsFlag(t, &opts.Help, expected, args)
		}
	})
	t.Run("verbose", func(t *testing.T) {
		verbose := map[string]bool{"-v": true, "-verbose": true, "--verbose": true, "": false, "-dry-run": false}
		for arg, expected := range verbose {
			opts.Verbose = false
			args := []string{arg}
			assertOptionsParse(t, opts, args)
			assertOptionsFlag(t, &opts.Verbose, expected, args)
		}
	})
	t.Run("version", func(t *testing.T) {
		version := map[string]bool{"-version": true, "--version": true, "": false, "-v": false}
		for arg, expected := range version {
			opts.Version = false
			args := []string{arg}
			assertOptionsParse(t, opts, args)
			assertOptionsFlag(t, &opts.Version, expected, args)
		}
	})
	t.Run("dry-run", func(t *testing.T) {
		dryRun := map[string]bool{"-dry-run": true, "--dry-run": true, "": false, "-v": false}
		for arg, expected := range dryRun {
			opts.DryRun = false
			args := []string{arg}
			assertOptionsParse(t, opts, args)
			assertOptionsFlag(t, &opts.DryRun, expected, args)
		}
	})
	t.Run("log-file", func(t *testing.T) {
		logFile := map[string]string{"-log-file": "foo", "--log-file": "bar", "--dry-run": "", "": ""}
		for arg, expected := range logFile {
			opts.LogFile = ""
			args := []string{arg, expected}
			assertOptionsParse(t, opts, args)
			assertOptionsFlag(t, &opts.LogFile, expected, args)
		}
	})
	t.Run("log-level", func(t *testing.T) {
		opts.LogLevel = LogLevelFatal
		levels := []string{"fatal", "debug", "info", "warning", "error", "fatal"}
		for _, level := range levels {
			args := []string{"-log-level", level}
			assertOptionsParse(t, opts, args)
			expected, err := ParseLogLevel(level)
			if err != nil {
				t.Errorf("ParseLogLevel(%q) failed: %v", level, err)
				continue
			}
			assertOptionsFlag(t, &opts.LogLevel, expected, args)
		}
	})
}
