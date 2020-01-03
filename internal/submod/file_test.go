package submod

import (
	"bytes"
	"strings"
	"testing"

	"golang.org/x/mod/modfile"
)

var testModData = []byte(`module example.com/a

replace example.com/a/b => ./b
replace example.com/a/c/v2 => ./c

require (
	example.com/a/b v1.0.0
	example.com/a/c/v2 v2.0.0
)
`)

var testModDataWithoutReplaces = []byte(`module example.com/a

require (
	example.com/a/b v1.0.0
	example.com/a/c/v2 v2.0.0
)
`)

var testModDataWithReplaces = []byte(`module example.com/a

require (
	example.com/a/b v1.0.0
	example.com/a/c/v2 v2.0.0
)

replace example.com/a/b => ./b

replace example.com/a/c/v2 => ./c
`)

func TestFile_AddSubmoduleReplaces(t *testing.T) {
	modFile, err := modfile.Parse("", testModDataWithoutReplaces, nil)
	if err != nil {
		t.Fatalf("failed to parse data: %v", err)
	}

	f := File{File: modFile}
	err = f.AddSubmoduleReplaces()
	if err != nil {
		t.Fatalf("failed to remove replaces: %v", err)
	}
	data, err := f.Format()
	if err != nil {
		t.Fatalf("failed to format mod file: %v", err)
	}

	if !bytes.Equal(testModDataWithReplaces, data) {
		t.Fatalf("expected %s but got %s", testModDataWithReplaces, data)
	}
}

func TestFile_AddSubmoduleReplacesWithModules(t *testing.T) {
	modFile, err := modfile.Parse("", testModDataWithoutReplaces, nil)
	if err != nil {
		t.Fatalf("failed to parse data: %v", err)
	}

	f := File{File: modFile}
	err = f.AddSubmoduleReplaces("example.com/a/c/v2")
	if err != nil {
		t.Fatalf("failed to remove replaces: %v", err)
	}
	data, err := f.Format()
	if err != nil {
		t.Fatalf("failed to format mod file: %v", err)
	}

	expected := "replace example.com/a/c/v2 => ./c"
	if !strings.Contains(string(data), expected) {
		t.Fatalf("expected %s to contain %s", data, expected)
	}
	unexpected := "replace example.com/a/b => ./b"
	if strings.Contains(string(data), unexpected) {
		t.Fatalf("unexpected %s in %s", data, unexpected)
	}
}

func TestFile_RemoveSubmoduleReplaces(t *testing.T) {
	modFile, err := modfile.Parse("", testModData, nil)
	if err != nil {
		t.Fatalf("failed to parse data: %v", err)
	}

	f := File{File: modFile}
	err = f.RemoveSubmoduleReplaces()
	if err != nil {
		t.Fatalf("failed to remove replaces: %v", err)
	}
	data, err := f.Format()
	if err != nil {
		t.Fatalf("failed to format mod file: %v", err)
	}

	if !bytes.Equal(testModDataWithoutReplaces, data) {
		t.Fatalf("expected %s but got %s", testModDataWithoutReplaces, data)
	}
}

func TestFile_RemoveSubmoduleReplacesWithModules(t *testing.T) {
	modFile, err := modfile.Parse("", testModData, nil)
	if err != nil {
		t.Fatalf("failed to parse data: %v", err)
	}

	f := File{File: modFile}
	err = f.RemoveSubmoduleReplaces("example.com/a/c/v2")
	if err != nil {
		t.Fatalf("failed to remove replaces: %v", err)
	}
	data, err := f.Format()
	if err != nil {
		t.Fatalf("failed to format mod file: %v", err)
	}

	absent := "replace example.com/a/c/v2 => ./c"
	if strings.Contains(string(data), absent) {
		t.Fatalf("expected %s to not contain %s", data, absent)
	}
	expected := "replace example.com/a/b => ./b"
	if !strings.Contains(string(data), expected) {
		t.Fatalf("expected %s in %s", data, expected)
	}
}
