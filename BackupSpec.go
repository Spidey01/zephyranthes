// SPDX-License-Identifier: Zlib
// Copyright 2024-2026, Terry M. Poulin.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Defines the individual backup job and resulting archive.
type BackupSpec struct {
	// Name of the backup job for debugging.
	Name string `yaml:"name" json:"name"`
	// Path to the output archive
	Path string `yaml:"path" json:"path"`
	// Which format to use for path.
	Format string `yaml:"format" json:"format"`
	// What to stuff in the archive.
	Contents []string `yaml:"contents" json:"contents"`
	// Commands to run before contents are added to the archive.
	Before CommandSpec `yaml:"before" json:"before"`
	// Commands to run after contents are added to the archive.
	After CommandSpec `yaml:"after" json:"after"`
}

// Defines how to execute external commands as part of the backup process.
type CommandSpec struct {
	// Overrides the working directory for the command.
	Directory string `yaml:"directory" json:"directory"`
	// Overrides the environment inherited from os.Environ.
	Environment map[string]string `yaml:"environment" json:"environment"`
	// Defines commands to be run before the backup contents are archived.
	Commands []string `yaml:"commands" json:"commands"`
	// Timeout is the execution deadline for running Commands.
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
}

const (
	FormatTGZ    = "tgz"
	FormatTar    = "tar"
	FormatTarGz  = "tar.gz"
	FormatTZST   = "tzst"
	FormatTarZst = "tar.zst"
	FormatZip    = "zip"
)

func UnmarshalBackupSpecs(data []byte) ([]BackupSpec, error) {
	var err error
	var prevErr error
	var spec []BackupSpec
	err = json.Unmarshal(data, &spec)
	if err != nil {
		prevErr = fmt.Errorf("error parsing JSON: %w", err)
	} else {
		return spec, nil
	}
	spec = []BackupSpec{}
	err = yaml.Unmarshal(data, &spec)
	if err != nil {
		spec = []BackupSpec{}
		return nil, fmt.Errorf("%w\nerror parsing YAML: %w", prevErr, err)
	}
	return spec, nil
}

func BackupSpecsFromFile(name string) ([]BackupSpec, error) {
	reader, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return UnmarshalBackupSpecs(data)
}
