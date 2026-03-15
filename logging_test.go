// SPDX-License-Identifier: Zlib
// Copyright 2026, Terry M. Poulin.
package main

import (
	"context"
	"os"
	"testing"
)

func TestSetupLogging(t *testing.T) {
	resetGlobals := func() {
		logger = nil
		logLevel = LogLevelFatal
		logPrefix = ""
	}
	defer resetGlobals()
	assert := func(ctx context.Context, prefix string, level LogLevel, logFile string) bool {
		if err := SetupLogging(ctx, prefix, level, logFile); err != nil {
			t.Fatalf("SetupLogging() failed: %v", err)
			return true
		}
		if logPrefix != prefix {
			t.Errorf("logPrefix %q, but expected %q", logPrefix, prefix)
			return true
		}
		if logger == nil && logFile != "" {
			t.Errorf("SetupLogging() didn't fail, but logger is nil")
			return true
		}
		return false
	}
	t.Run("nologfile", func(t *testing.T) {
		resetGlobals()
		assert(t.Context(), t.Name(), LogLevelDebug, "")
		if logger != nil {
			t.Errorf("logger set, but no log file specified")
		}
	})
	t.Run("logfile", func(t *testing.T) {
		resetGlobals()
		fp, err := os.CreateTemp("", "")
		if err != nil {
			t.Fatalf("os.CreateTemp() failed: %v", err)
		}
		defer fp.Close()
		defer os.Remove(fp.Name())
		assert(t.Context(), t.Name(), LogLevelDebug, fp.Name())
	})
	t.Run("stdin", func(t *testing.T) {
		resetGlobals()
		assert(t.Context(), t.Name(), LogLevelDebug, "-")
		if logger.Writer() != os.Stdout {
			t.Errorf("logger output is not os.Stdout")
		}
	})
}

func TestParseLogLevel(t *testing.T) {
	// input -> expected output
	values := map[string]LogLevel{
		// Parsing from lower case.
		"debug":   LogLevelDebug,
		"verbose": LogLevelVerbose,
		"info":    LogLevelInfo,
		"warning": LogLevelWarning,
		"error":   LogLevelError,
		"fatal":   LogLevelFatal,
		// Parsing from upper case.
		"DEBUG":   LogLevelDebug,
		"VERBOSE": LogLevelVerbose,
		"INFO":    LogLevelInfo,
		"WARNING": LogLevelWarning,
		"ERROR":   LogLevelError,
		"FATAL":   LogLevelFatal,
		// Parsing silly cases.
		"DebUG":   LogLevelDebug,
		"VERboSE": LogLevelVerbose,
		"iNFo":    LogLevelInfo,
		"WARNing": LogLevelWarning,
		"ErrOR":   LogLevelError,
		"FaTaL":   LogLevelFatal,
	}
	for input, expected := range values {
		actual, err := ParseLogLevel(input)
		if err != nil {
			t.Errorf("ParseLogLevel(%q) -> error: %v", input, err)
			continue
		}
		if actual != expected {
			t.Errorf("ParseLogLevel(%q) -> actual: %s expected: %v", input, actual, expected)
		}
	}
	if level, err := ParseLogLevel("invalid"); err == nil {
		t.Errorf("ParseLogLevel(%q) -> returned %s instead of error", "invalid", level)
	}
	if level, err := ParseLogLevel(""); err == nil {
		t.Errorf("ParseLogLevel(%q) -> returned %s instead of error", "", level)
	}
}
