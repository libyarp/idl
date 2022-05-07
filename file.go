package idl

import "path/filepath"

// File represents a single YARP source file.
type File struct {
	// Tree contains a list of Package, Import, Message, and Service objects
	// representing structures defined in a source file.
	Tree []any

	// Package represents the package name defined by the source file.
	Package string

	// DeclaredMessages contains the names of all messages declared by the
	// source file.
	DeclaredMessages []string

	// DeclaredService contains the names of all services declared by the
	// source file.
	DeclaredServices []string

	// ImportedFiles contains a list of paths provided to `import` directives.
	ImportedFiles []string
	declaredNames map[string]any
}

func (f *File) push(val any) {
	f.Tree = append(f.Tree, val)
	switch v := val.(type) {
	case Package:
		f.Package = v.Name
	case Import:
		f.ImportedFiles = append(f.ImportedFiles, filepath.Clean(v.Path))
	case Message:
		f.DeclaredMessages = append(f.DeclaredMessages, v.Name)
		if f.declaredNames == nil {
			f.declaredNames = map[string]any{}
		}
		f.declaredNames[v.Name] = &v
	case Service:
		f.DeclaredServices = append(f.DeclaredServices, v.Name)
		if f.declaredNames == nil {
			f.declaredNames = map[string]any{}
		}
		f.declaredNames[v.Name] = &v
	}
}

func (f *File) isImported(path string) bool {
	path = filepath.Clean(path)
	for _, p := range f.ImportedFiles {
		if p == path {
			return true
		}
	}
	return false
}

func (f *File) isDefined(name string) bool {
	if f.declaredNames == nil {
		return false
	}
	_, ok := f.declaredNames[name]
	return ok
}

func (f *File) definitionByName(name string) any {
	if f.declaredNames == nil {
		return nil
	}
	return f.declaredNames[name]
}

func (f File) last() any {
	return f.Tree[len(f.Tree)-1]
}

// MessageByName takes a name and returns a Message, along with a boolean
// indicating whether the provided message exists in the current File.
func (f File) MessageByName(name string) (*Message, bool) {
	v, ok := f.declaredNames[name]
	if !ok {
		return nil, false
	}
	m, ok := v.(*Message)
	return m, ok
}

// ServiceByName takes a name and returns a Service, along with a boolean
// indicating whether the provided service exists in the current File.
func (f File) ServiceByName(name string) (*Service, bool) {
	v, ok := f.declaredNames[name]
	if !ok {
		return nil, false
	}
	s, ok := v.(*Service)
	return s, ok
}
