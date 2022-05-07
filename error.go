package idl

import (
	"fmt"
)

// ParseError indicates that one or more productions from the scanner does not
// define a valid IDL file.
type ParseError struct {
	Token   Token
	Message string
}

func (p ParseError) Error() string {
	return fmt.Sprintf("%s at %#v on line %d, column %d", p.Message, p.Token.Value, p.Token.Line, p.Token.Column)
}

// SyntaxError indicates that a provided file does not contain a valid YARP
// Interface Description File.
type SyntaxError struct {
	Message string
	Line    int
	Column  int
}

func (s SyntaxError) Error() string {
	return fmt.Sprintf("%s at line %d, column %d", s.Message, s.Line, s.Column)
}

// SourceFileNotFoundError indicates that the process could not find an input
// file provided by the user.
type SourceFileNotFoundError struct{ Path string }

func (s SourceFileNotFoundError) Error() string { return fmt.Sprintf("%s: no such file", s.Path) }

// ImportFileNotFoundError indicates that a file imported by a given source
// could not be found.
type ImportFileNotFoundError struct{ Source, Path string }

func (i ImportFileNotFoundError) Error() string {
	return fmt.Sprintf("%s (imported by %s): no such file", i.Source, i.Path)
}

// SourceIsDirectoryError indicates that a given source file is, in fact, a
// directory, and therefore cannot be read.
type SourceIsDirectoryError struct{ Path string }

func (s SourceIsDirectoryError) Error() string { return fmt.Sprintf("%s: is a directory", s.Path) }

// MixedPackagesError indicates that source files provides different packages.
// Only a single package can be compiled at a time.
type MixedPackagesError struct{ Path, Package1, Package2 string }

func (m MixedPackagesError) Error() string {
	return fmt.Sprintf("mixed packages in source (reading %s): found both %s and %s", m.Path, m.Package1, m.Package2)
}
