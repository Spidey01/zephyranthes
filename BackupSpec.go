// SPDX-License-Identifier: Zlib
// Copyright 2024, Terry M. Poulin.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Defines the individual backup job and resulting archive.
type BackupSpec struct {
	Name     string   `yaml:"name" json:"name"`
	Path     string   `yaml:"path" json:"path"`
	Format   string   `yaml:"format" json:"format"`
	Contents []string `yaml:"contents" json:"contents"`
}

const (
	FormatTar = "tar"
	FormatZip = "zip"
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
