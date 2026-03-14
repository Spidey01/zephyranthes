// SPDX-License-Identifier: Zlib
// Copyright 2026, Terry M. Poulin.

package main

import (
	"encoding/json"
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestUnmarshalBackupSpecs(t *testing.T) {
	expected := []BackupSpec{
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
	eq := func(e1, e2 BackupSpec) bool {
		return e1.Name == e2.Name &&
			e1.Path == e2.Path &&
			e1.Format == e2.Format &&
			slices.Equal(e1.Contents, e2.Contents)
	}
	t.Run("JSON", func(t *testing.T) {
		data, err := json.MarshalIndent(expected, "", "    ")
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		actual, err := UnmarshalBackupSpecs(data)
		if err != nil {
			t.Fatalf("UnmarshalBackupSpecs() from JSON failed: %v", err)
		}
		t.Logf("actual: %v", actual)
		if !slices.EqualFunc(actual, expected, eq) {
			t.Errorf("specs do not match:\nactual  : %+v\nexpected: %+v\n", actual, expected)
		}
	})
	t.Run("YAML", func(t *testing.T) {
		data, err := yaml.Marshal(expected)
		if err != nil {
			t.Fatalf("yaml.Marshal failed: %v", err)
		}
		actual, err := UnmarshalBackupSpecs(data)
		if err != nil {
			t.Fatalf("UnmarshalBackupSpecs() from YAML failed: %v", err)
		}
		t.Logf("actual: %v", actual)
		if !slices.EqualFunc(actual, expected, eq) {
			t.Errorf("specs do not match:\nactual  : %+v\nexpected: %+v\n", actual, expected)
		}
	})
}
