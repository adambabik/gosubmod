package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/adambabik/gosubmod/internal/submod"
	"golang.org/x/mod/modfile"
)

var usage = `gosubmod is a tool that simplifies working with Go submodules.

Usage:

	gosubmod <command> [arguments] submodules...

The commands are:

	list    list all the recognized submodules
	add     add "replace" directives with relative paths for submodules
	drop    drop "replace" directives with relative paths for submodules
	bump    bump updates submodule versions; requires a semver part (-patch, -minor or -major) as an argument

`

type command func(*submod.File) error

func main() {
	var cmd command
	// interpret provided arguments
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "list", "l":
			cmd = listCmd
		case "add", "a":
			cmd = addCmd
		case "drop", "d":
			cmd = dropCmd
		}
	}
	// if a command has not been found, print usage
	if cmd == nil {
		usageCmd()
		os.Exit(2)
	}
	// read and parse go mod file
	modFilePath, err := goModFilePath()
	if err != nil {
		log.Fatalf("failed to create a go.mod file path: %v", err)
	}
	modFile, err := readModFile(modFilePath)
	if err != nil {
		log.Fatalf("failed to read mod file: %v", err)
	}
	f := &submod.File{
		File:     modFile,
		Filepath: modFilePath,
		Strict:   true,
	}
	if err := cmd(f); err != nil {
		log.Fatalf("failed to execute command: %v", err)
	}
}

func usageCmd() {
	fmt.Fprint(os.Stderr, usage)
}

func listCmd(f *submod.File) error {
	submodules := f.Submodules()
	w := bufio.NewWriter(os.Stdout)
	for _, m := range submodules {
		_, err := fmt.Fprintf(w, "%s\n", m)
		if err != nil {
			return err
		}
	}
	return w.Flush()
}

func addCmd(f *submod.File) error {
	var modules []string
	if len(os.Args) > 2 {
		modules = os.Args[2:]
	}
	err := f.AddSubmoduleReplaces(modules...)
	if err != nil {
		return err
	}
	return writeModFile(f)
}

func dropCmd(f *submod.File) error {
	var modules []string
	if len(os.Args) > 2 {
		modules = os.Args[2:]
	}
	err := f.RemoveSubmoduleReplaces(modules...)
	if err != nil {
		return err
	}
	return writeModFile(f)
}

func goModFilePath() (string, error) {
	// filepath.Abs detects the current working directory.
	return filepath.Abs("go.mod")
}

func readModFile(path string) (*modfile.File, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return modfile.Parse(path, data, nil)
}

func writeModFile(f *submod.File) error {
	f.Cleanup()
	data, err := f.Format()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(f.Filepath, data, 0755)
}
