// SPDX-License-Identifier: Zlib
// Copyright 2026, Terry M. Poulin.

package main

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
)

var testBackupSpecs = []BackupSpec{
	{
		Name:     "firstspec",
		Path:     "firstpath",
		Format:   "firstformat",
		Contents: []string{"foo", "bar"},
	},
	{
		Name:     "secondspec",
		Path:     "secondpath",
		Format:   "secondformat",
		Contents: []string{"ham", "spam", "eggs"},
	},
}

func assertBackupSpecs(t *testing.T, actual, expected []BackupSpec) bool {
	eq := func(e1, e2 BackupSpec) bool {
		return e1.Name == e2.Name &&
			e1.Path == e2.Path &&
			e1.Format == e2.Format &&
			slices.Equal(e1.Contents, e2.Contents)
	}
	if !slices.EqualFunc(actual, expected, eq) {
		t.Errorf("specs do not match:\nactual  : %+v\nexpected: %+v\n", actual, expected)
		return true
	}
	return false
}

func TestUnmarshalBackupSpecs(t *testing.T) {
	assert := func(t *testing.T, expected []BackupSpec, marshaler func(any) ([]byte, error)) {
		data, err := json.MarshalIndent(expected, "", "    ")
		if err != nil {
			t.Fatalf("marshaling failed: %v", err)
		}
		actual, err := UnmarshalBackupSpecs(data)
		if err != nil {
			t.Fatalf("UnmarshalBackupSpecs() failed: %v", err)
		}
		assertBackupSpecs(t, actual, expected)
	}
	t.Run("JSON", func(t *testing.T) {
		assert(t, testBackupSpecs, json.Marshal)
	})
	t.Run("YAML", func(t *testing.T) {
		assert(t, testBackupSpecs, yaml.Marshal)
	})
}

func TestBackupSpecsFromFile(t *testing.T) {
	assert := func(t *testing.T, expected []BackupSpec, marshaler func(any) ([]byte, error)) {
		fp, err := os.CreateTemp("", "")
		if err != nil {
			t.Fatalf("os.CreateTemp() failed: %v", err)
		}
		name := fp.Name()
		defer os.Remove(name)
		data, err := marshaler(expected)
		if err != nil {
			t.Fatalf("marshaling failed: %v", err)
		}
		if _, err = fp.Write(data); err != nil {
			t.Fatalf("Writing marshaled data failed: %v", err)
		}
		fp.Close()
		actual, err := BackupSpecsFromFile(name)
		if err != nil {
			t.Errorf("BackupSpecsFromFile(%q) failed: %v", name, err)
		}
		assertBackupSpecs(t, actual, expected)
	}
	t.Run("JSON", func(t *testing.T) {
		assert(t, testBackupSpecs, json.Marshal)
	})
	t.Run("YAML", func(t *testing.T) {
		assert(t, testBackupSpecs, yaml.Marshal)
	})
}
