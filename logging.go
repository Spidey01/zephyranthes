// SPDX-License-Identifier: Zlib
// Copyright 2024-2025, Terry M. Poulin.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type LogLevel int

const (
	LogLevelFatal LogLevel = iota
	LogLevelError
	LogLevelWarning
	LogLevelInfo
	LogLevelVerbose
	LogLevelDebug
)

var logger *log.Logger
var logLevel LogLevel
var logPrefix string

func (ll LogLevel) String() string {
	switch ll {
	case LogLevelFatal:
		return "FATAL"
	case LogLevelError:
		return "ERROR"
	case LogLevelWarning:
		return "WARNING"
	case LogLevelInfo:
		return "INFO"
	case LogLevelVerbose:
		return "VERBOSE"
	case LogLevelDebug:
		return "DEBUG"
	default:
		return ""
	}
}

func ParseLogLevel(arg string) (ll LogLevel, err error) {
	switch strings.ToUpper(arg) {
	case LogLevelFatal.String():
		ll = LogLevelFatal
	case LogLevelError.String():
		ll = LogLevelError
	case LogLevelWarning.String():
		ll = LogLevelWarning
	case LogLevelInfo.String():
		ll = LogLevelInfo
	case LogLevelVerbose.String():
		ll = LogLevelVerbose
	case LogLevelDebug.String():
		ll = LogLevelDebug
	default:
		err = fmt.Errorf("invalid log level: %s", arg)
	}
	return
}

// Initializes the log level and sets up a logger for the specified file.
func SetupLogging(prefix string, level LogLevel, logFile string) error {
	logPrefix = prefix
	logLevel = level
	if logFile != "" {
		fp, err := os.Create(logFile)
		if err != nil {
			return err
		}
		logger = log.New(fp, logPrefix+" ", log.LstdFlags|log.Lmsgprefix)
		context.AfterFunc(context.Background(), func() { fp.Close() })
	}
	return nil
}

// Writes the formatted message to the active log file with the given prefix.
// Used by the various log functions to set a log level like prefix.
func LogMsg(prefix LogLevel, format string, args ...any) {
	if logger != nil && prefix <= logLevel {
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

func Debugf(format string, args ...any) {
	LogMsg(LogLevelDebug, format, args...)
}

func Verbosef(format string, args ...any) {
	if options.Verbose {
		FmtMsg(os.Stdout, "", format, args...)
	}
	LogMsg(LogLevelVerbose, format, args...)
}

func Infof(format string, args ...any) {
	FmtMsg(os.Stdout, "", format, args...)
	LogMsg(LogLevelInfo, format, args...)
}

func Warningf(format string, args ...any) {
	ErrMsg("WARNING:", format, args...)
	LogMsg(LogLevelWarning, format, args...)
}

func Errorf(format string, args ...any) {
	ErrMsg("ERROR:", format, args...)
	LogMsg(LogLevelError, format, args...)
}

func Fatalf(format string, args ...any) {
	ErrMsg("FATAL:", format, args...)
	LogMsg(LogLevelFatal, format, args...)
	os.Exit(1)
}

// Like Fatalf, but the prefix for stderr is the program name.
func Die(format string, args ...any) {
	ErrMsg(options.Name(), format, args...)
	LogMsg(LogLevelFatal, format, args...)
	os.Exit(1)
}
