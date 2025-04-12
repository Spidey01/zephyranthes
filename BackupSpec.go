// SPDX-License-Identifier: Zlib
// Copyright 2024, Terry M. Poulin.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
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

// Executes the backup specification using the provided context. Returns nil
// once the job is complete, or an error is the operation failed.
func (backup *BackupSpec) Do(ctx context.Context) error {
	archive, err := CreateArchive(backup.Path, backup.Format)
	if err != nil {
		return err
	}
	defer archive.Close()
	Verbosef("Archiving contents...")
	for _, fn := range backup.Contents {
		stat, err := os.Stat(fn)
		if err != nil {
			Warningf("Skipping %s: %v", fn, err)
			continue
		}
		Verbosef("Inspecting %s", fs.FormatFileInfo(stat))
		if stat.IsDir() {
			Infof("Adding directory tree %s", stat.Name())
			err = archive.AddDir(stat.Name(), fn)
		} else {
			Infof("Adding file %s", stat.Name())
			err = archive.AddFile(stat.Name(), fn)
		}
		if err != nil {
			Errorf("Failed to backup %s: %v", fn, err)
			break
		}
	}
	return nil
}
