// SPDX-License-Identifier: Zlib
// Copyright 2024-2025, Terry M. Poulin.
package main

import "context"

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
			err = spec.Do(context.Background())
			if err != nil {
				Fatalf("Backup %s failed: %v", spec.Name, err)
			}
		}
	}
}
