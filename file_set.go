package idl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileSet represents structures provided by a set of source files.
type FileSet struct {
	loadedFiles   map[string]bool
	knownServices map[string]bool
	packageName   string
	messages      map[string]*Message
	Messages      []*Message
	Services      []*Service
}

// NewFileSet creates a new FileSet structure
func NewFileSet() *FileSet {
	return &FileSet{
		loadedFiles:   map[string]bool{},
		knownServices: map[string]bool{},
		packageName:   "",
		messages:      map[string]*Message{},
		Messages:      nil,
		Services:      nil,
	}
}

func (f *FileSet) registerMessage(file *File, msg *Message) error {
	fqn := fmt.Sprintf("%s.%s", file.Package, msg.Name)
	if f.messages == nil {
		f.messages = map[string]*Message{}
	}
	if _, ok := f.messages[fqn]; ok {
		// TODO: Normalize errors
		return fmt.Errorf("duplicated definition of %s", fqn)
	}
	f.messages[fqn] = msg
	return nil
}

func (f FileSet) isLoaded(path string) bool {
	_, ok := f.loadedFiles[path]
	return ok
}

func (f FileSet) findAndLoad(path string) (string, *File, error) {
	s, err := filepath.Abs(path)
	if err != nil {
		return "", nil, err
	}
	if f.isLoaded(s) {
		return "", nil, nil
	}
	stat, err := os.Stat(s)
	exist := true
	if os.IsNotExist(err) {
		exist = false
	} else if err != nil {
		return "", nil, err
	}

	if !exist || stat.IsDir() {
		next := path + ".yarp"
		st, err := os.Stat(next)
		if err == nil && !st.IsDir() {
			stat = st
			path = next
			exist = true
		}
	} else {
		path = s
	}

	if !exist {
		return "", nil, SourceFileNotFoundError{Path: path}
	}
	if stat.IsDir() {
		return "", nil, SourceIsDirectoryError{Path: path}
	}

	file, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	tokens, err := Scan(file)
	if err != nil {
		return "", nil, err
	}
	result, err := Parse(tokens)
	if err != nil {
		return "", nil, err
	}

	return path, result, nil
}

// Load attempts to load a given file under the provided path and add its
// contents to the current FileSet. In case the file cannot be loaded, an error
// is returned.
func (f *FileSet) Load(path string) error {
	finalPath, file, err := f.findAndLoad(path)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	f.loadedFiles[finalPath] = true
	if f.packageName == "" {
		f.packageName = file.Package
	} else if f.packageName != file.Package {
		return MixedPackagesError{
			Path:     path,
			Package1: f.packageName,
			Package2: file.Package,
		}
	}

	if err = f.processImports(finalPath, file); err != nil {
		return err
	}

	for _, n := range file.DeclaredMessages {
		m, ok := file.MessageByName(n)
		if !ok {
			return fmt.Errorf("BUG: %s declares %s, but message could not be found", finalPath, n)
		}
		f.Messages = append(f.Messages, m)
		if err = f.registerMessage(file, m); err != nil {
			return err
		}
	}

	for _, n := range file.DeclaredServices {
		s, ok := file.ServiceByName(n)
		if !ok {
			return fmt.Errorf("BUG: %s declares %s, but service could not be found", finalPath, n)
		}
		if f.knownServices == nil {
			f.knownServices = map[string]bool{}
		}
		if _, ok := f.knownServices[n]; ok {
			return fmt.Errorf("multiple declarations of service %s (duplicate found in %s)", n, finalPath)
		}
		f.knownServices[n] = true
		f.Services = append(f.Services, s)
	}
	return nil
}

func (f *FileSet) processImports(path string, file *File) error {
	for _, i := range file.ImportedFiles {
		pwd := filepath.Dir(path)
		target, err := filepath.Abs(filepath.Join(pwd, i))
		if err != nil {
			return err
		}
		finalPath, imported, err := f.findAndLoad(target)
		if err != nil {
			if nf, ok := err.(SourceFileNotFoundError); ok {
				return ImportFileNotFoundError{
					Source: path,
					Path:   nf.Path,
				}
			} else {
				return err
			}
		}
		f.loadedFiles[finalPath] = true
		if err := f.processImports(finalPath, imported); err != nil {
			return err
		}
		for _, m := range imported.DeclaredMessages {
			msg, ok := imported.MessageByName(m)
			if !ok {
				return fmt.Errorf("BUG: %s declares %s, but message could not be found", finalPath, m)
			}
			if err = f.registerMessage(imported, msg); err != nil {
				return err
			}
			if imported.Package == f.packageName {
				f.Messages = append(f.Messages, msg)
			}
		}
		if imported.Package == f.packageName {
			for _, n := range imported.DeclaredServices {
				s, ok := imported.ServiceByName(n)
				if !ok {
					return fmt.Errorf("BUG: %s declares %s, but service could not be found", finalPath, n)
				}
				if f.knownServices == nil {
					f.knownServices = map[string]bool{}
				}
				if _, ok := f.knownServices[n]; ok {
					return fmt.Errorf("multiple declarations of service %s (duplicate found in %s)", n, finalPath)
				}
				f.knownServices[n] = true
				f.Services = append(f.Services, s)
			}
		}
	}
	return nil
}

// FindMessage takes a message name (e.g. SomethingRequest) or FQN (e.g.
// package.SomethingRequest) and returns a Message along with a boolean
// indicating whether the provided name could be resolved to a message.
func (f *FileSet) FindMessage(name string) (*Message, bool) {
	n := name
	if !strings.ContainsRune(n, '.') {
		// N should be present in the package we're processing.
		n = fmt.Sprintf("%s.%s", f.packageName, n)
	}

	m, ok := f.messages[n]
	return m, ok
}

// Package returns the package declared by loaded source files.
func (f FileSet) Package() string {
	return f.packageName
}

// FromSamePackage takes a name and returns whether it is declared by the
// package declared in the loaded source files.
func (f FileSet) FromSamePackage(name string) bool {
	n := name
	if !strings.ContainsRune(n, '.') {
		// N should be present in the package we're processing.
		n = fmt.Sprintf("%s.%s", f.packageName, n)
	}
	m, ok := f.messages[n]
	if !ok {
		return false
	}
	return strings.TrimSuffix(n, "."+m.Name) == f.packageName
}

// SplitComponents splits a given name into a package and message name. In case
// the provided name does not include a package, pkgName is returned as an empty
// string.
func SplitComponents(n string) (pkgName, messageName string) {
	components := strings.Split(n, ".")
	if len(components) == 1 {
		return "", components[0]
	}

	return strings.Join(components[0:len(components)-1], "."), components[len(components)-1]
}
