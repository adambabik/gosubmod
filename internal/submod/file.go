package submod

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

type FileError struct {
	Filepath string
	Modules  []string
	Err      error
}

func (e *FileError) Error() string {
	return fmt.Sprintf(`file error for path "%s" with modules %+v: %v`, e.Filepath, e.Modules, e.Err)
}

func (e *FileError) Unwrap() error {
	return e.Err
}

type Version = module.Version

// File is an extension over modfile.File which supports submodules.
// Submodule is defined as a nested module, i.e. having a parent directory
// which is a Go module.
type File struct {
	*modfile.File

	// Filepath is a path to a main go.mod file.
	Filepath string

	// Strict verifies if the submodule are in a file system.
	Strict bool
}

// Submodules returns a list of submodule paths.
func (f *File) Submodules() (result []string) {
	for _, s := range f.submodules() {
		result = append(result, s.Mod.String())
	}
	return
}

func (f *File) versions() (result []Version) {
	for _, s := range f.submodules() {
		result = append(result, s.Mod)
	}
	return
}

func (f *File) parseModules(modules []string) (modVersions, error) {
	if len(modules) == 0 {
		return f.versions(), nil
	}
	return parseModules(modules)
}

// AddSubmoduleReplaces adds "replace" directives for all detected submodules.
// All added replaces are relative paths pointing at the found child directories.
func (f *File) AddSubmoduleReplaces(modules ...string) error {
	versions, err := f.parseModules(modules)
	if err != nil {
		return &FileError{
			Filepath: f.Filepath,
			Modules:  modules,
			Err:      err,
		}
	}

	for _, submodule := range f.submodules() {
		if !versions.ContainsPath(submodule.Mod) {
			continue
		}

		name := f.submoduleDirName(submodule.Mod)

		if f.Strict {
			dir, err := filepath.Abs(name)
			if err != nil {
				return err
			}
			if info, err := os.Stat(dir); err != nil {
				return err
			} else if !info.IsDir() {
				return fmt.Errorf("expected %s to be a dir", dir)
			}
		}

		newPath := "." + string(os.PathSeparator) + name
		if err := f.AddReplace(submodule.Mod.Path, "", newPath, ""); err != nil {
			return &FileError{
				Err:      err,
				Modules:  modules,
				Filepath: f.Filepath,
			}
		}
	}
	return nil
}

// RemoveSubmoduleReplaces removes all "replace" directives
// related to the found submodules.
func (f *File) RemoveSubmoduleReplaces(modules ...string) error {
	versions, err := f.parseModules(modules)
	if err != nil {
		return &FileError{
			Filepath: f.Filepath,
			Modules:  modules,
			Err:      err,
		}
	}

	submodules := f.submodules()
	for _, r := range f.Replace {
		for _, s := range submodules {
			if !versions.ContainsPath(s.Mod) {
				continue
			}
			// TODO: what about versions match?
			if r.Old.Path == s.Mod.Path && strings.HasPrefix(r.New.Path, ".") {
				if err := f.DropReplace(r.Old.Path, r.Old.Version); err != nil {
					return &FileError{
						Err:      err,
						Modules:  modules,
						Filepath: f.Filepath,
					}
				}
			}
		}
	}
	return nil
}

// Format cleanups the parsed mod file and returns it as bytes.
func (f *File) Format() ([]byte, error) {
	f.File.Cleanup()
	return f.File.Format()
}

func (f *File) submodules() (result []*modfile.Require) {
	modPath := f.Module.Mod.Path
	for _, r := range f.Require {
		if isSubmodule(r.Mod.Path, modPath) {
			result = append(result, r)
		}
	}
	return
}

func (f *File) submoduleDirName(submodule module.Version) string {
	// Split prefix and major version. The major version is not
	// a part of the directory name.
	prefix, _, _ := module.SplitPathVersion(submodule.Path)
	// Remove the main module prefix.
	return strings.TrimPrefix(prefix, f.Module.Mod.Path+"/")
}

// isSubmodule returns true if the path has a chance to be a submodule of mainPath.
// TODO: this should be more strict and examine go.mod to verify it's a submodule.
func isSubmodule(path, mainPath string) bool {
	return strings.HasPrefix(path, mainPath)
}

type modVersions []module.Version

func (mv modVersions) ContainsPath(version module.Version) bool {
	for _, v := range mv {
		if v.Path == version.Path {
			return true
		}
	}
	return false
}

func parseModules(modules []string) (modVersions, error) {
	result := make([]module.Version, 0, len(modules))
	for _, path := range modules {
		if err := module.CheckPath(path); err != nil {
			return nil, err
		}
		// After module.CheckPath we know that module.SplitPathVersion will succeed.
		prefix, version, _ := module.SplitPathVersion(path)
		result = append(result, module.Version{
			Path:    prefix + version, // join the prefix and version again to form a valid path
			Version: module.CanonicalVersion(version),
		})
	}
	return result, nil
}
