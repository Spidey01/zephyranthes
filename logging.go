// SPDX-License-Identifier: Zlib
// Copyright 2024, Terry M. Poulin.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var logger *log.Logger

func setupLogging(logFile string) error {
	if logFile != "" {
		fp, err := os.Create(logFile)
		if err != nil {
			return err
		}
		logger = log.New(fp, options.Name()+" ", log.LstdFlags|log.Lmsgprefix)
		context.AfterFunc(context.Background(), func() { fp.Close() })
	}
	return nil
}

// Writes the formatted message to the active log file with the given prefix.
// Used by the various log functions to set a log level like prefix.
func LogMsg(prefix, format string, args ...any) {
	if logger != nil {
		logger.Println(prefix, fmt.Sprintf(format, args...))
	}
}

// Like LogMsg but writes to output stream. If format does not end in a new
// line, one will be added. if prefix does not end in space, one will be
// inserted.
func FmtMsg(w io.Writer, prefix, format string, args ...any) {
	if prefix != "" {
		fmt.Fprint(w, prefix)
		if !strings.HasSuffix(format, " ") {
			fmt.Fprint(w, " ")
		}
	}
	fmt.Fprintf(os.Stderr, format, args...)
	if !strings.HasSuffix(format, "\n") {
		fmt.Fprint(w, "\n")
	}
}

// Does a FmtMsg to stderr.
func ErrMsg(prefix, format string, args ...any) {
	FmtMsg(os.Stderr, prefix, format, args...)
}

func Verbosef(format string, args ...any) {
	if options.Verbose {
		FmtMsg(os.Stdout, "", format, args...)
		LogMsg("VERBOSE:", format, args...)
	}
}

func Infof(format string, args ...any) {
	FmtMsg(os.Stdout, "", format, args...)
	LogMsg("INFO:", format, args...)
}

func Warningf(format string, args ...any) {
	ErrMsg("WARNING:", format, args...)
	LogMsg("WARNING:", format, args...)
}

func Errorf(format string, args ...any) {
	ErrMsg("ERROR:", format, args...)
	LogMsg("ERROR:", format, args...)
}

func Fatalf(format string, args ...any) {
	ErrMsg("FATAL:", format, args...)
	LogMsg("FATAL:", format, args...)
	os.Exit(1)
}

// Like Fatalf, but the prefix for stderr is the program name.
func Die(format string, args ...any) {
	ErrMsg(options.Name(), format, args...)
	LogMsg("FATAL:", format, args...)
	os.Exit(1)
}
