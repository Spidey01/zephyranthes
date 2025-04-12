// SPDX-License-Identifier: Zlib
// Copyright 2024-2025, Terry M. Poulin.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
)

type Options struct {
	// Verbose output.
	Verbose bool
	// Display help output.
	Help bool
	// Log details to this path.
	LogFile string
	// How verbose to make the log.
	LogLevel LogLevel
	// Perform a dry run.
	DryRun bool
	// Flag set for parsing the above options.
	FlagSet *flag.FlagSet
}

// Returns a new Options set to defaults. Call one of the parse functions to
// populate it.
func NewOptions() *Options {
	var opts Options
	fs := flag.NewFlagSet(opts.Name(), flag.ExitOnError)
	fs.BoolVar(&opts.Help, "h", false, "Show usage.")
	fs.BoolVar(&opts.Help, "help", false, "Show usage.")
	fs.BoolVar(&opts.Verbose, "v", false, "Produce verbose output.")
	fs.BoolVar(&opts.Verbose, "verbose", false, "Produce verbose output.")
	fs.StringVar(&opts.LogFile, "log-file", "", "Log what we're doing to the specified FILE.")
	fs.Func("log-level", "How verbose the log file is. One of: fatal, error, warning, info, verbose, debug", func(arg string) error {
		var err error
		opts.LogLevel, err = ParseLogLevel(arg)
		return err
	})
	fs.BoolVar(&opts.DryRun, "dry-run", false, "")
	fs.Usage = func() {
		out := fs.Output()
		io.WriteString(out, fmt.Sprintf("usage: %s [options] [file ...]\n", opts.Name()))
		io.WriteString(out, "\nOptions:\n\n")
		fs.PrintDefaults()
		io.WriteString(out, "\nEach file is parsed to define the backup archive(s) to create. Defaults to reading from standard input.\n")
	}
	opts.FlagSet = fs
	return &opts
}

// Parses command line options from os.Args, exiting if help requested or an
// error resulted.
func (opt *Options) MustParseArgs() {
	err := opt.ParseArgs()
	if err != nil {
		opt.ExitUsageError(err)
	} else if opt.Help {
		opt.ExitUsage(0)
	}
}

// Calls Parse() with os.Args, skipping the program name.
func (opt *Options) ParseArgs() error {
	return opt.Parse(os.Args[1:])
}

// Parse the specified args using our FlagSet.
func (opt *Options) Parse(args []string) error {
	return opt.FlagSet.Parse(args)
}

// Prints usage information and exits with status.
func (opt *Options) ExitUsage(status int) {
	opt.FlagSet.Usage()
	os.Exit(status)
}

// Prints error/usage information and exits with a none zero status.
func (opt *Options) ExitUsageError(err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", opt.Name(), err)
	opt.ExitUsage(64) // EX_USAGE.
}

// Returns the simple name of the program.
func (opt *Options) Name() string {
	return path.Base(os.Args[0])
}

// Returns remaining arguments after parsing.
func (opt *Options) Args() []string {
	return opt.FlagSet.Args()
}
