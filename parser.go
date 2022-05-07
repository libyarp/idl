package idl

import (
	"fmt"
	"strconv"
	"strings"
)

// Position represents a given Line/Column position within a source file.
type Position struct {
	Line   int
	Column int
}

// Offset represents the offset in which a given structure appears in the source
// file. It includes Position for both the point in which it starts, and the
// point in which it ends.
type Offset struct {
	StartsAt Position
	EndsAt   Position
}

// Package represents a `package` declaration in a source file.
type Package struct {
	Offset Offset
	Name   string
}

// Import represents a `import` statement, which includes a path to be loaded.
type Import struct {
	Offset Offset
	Path   string
}

// Message represents a single `message` declared in a source file.
type Message struct {
	Offset      Offset
	Name        string
	Comments    []string
	Annotations AnnotationCollection
	Fields      []any
}

// Service represents a single `service` declared in a source file.
type Service struct {
	Offset      Offset
	Name        string
	Comments    []string
	Annotations AnnotationCollection
	Methods     []Method
}

// AnnotationValue represents a single @annotation value present in a source
// file.
type AnnotationValue struct {
	Offset Offset
	Name   string
	Value  []string
}

const (
	// RepeatedAnnotation contains a constant representing the name of
	// @repeated annotations.
	RepeatedAnnotation = "repeated"

	// OptionalAnnotation contains a constant representing the name of
	// @optional annotations.
	OptionalAnnotation = "optional"

	// DeprecatedAnnotation contains a constant representing the name of
	// @deprecated annotations. (RFU)
	DeprecatedAnnotation = "deprecated"
)

// AnnotationCollection represents a list of Annotation values.
type AnnotationCollection []AnnotationValue

// FindByName takes a name and returns an Annotation value, and a boolean
// indicating whether the current AnnotationCollection contains the provided
// name.
func (a AnnotationCollection) FindByName(name string) (*AnnotationValue, bool) {
	for _, v := range a {
		if v.Name == name {
			return &v, true
		}
	}
	return nil, false
}

// Method represents a Service's method
type Method struct {
	Offset          Offset
	Name            string
	Comments        []string
	Annotations     AnnotationCollection
	ArgumentType    string
	ReturnType      string
	ReturnStreaming bool
}

// Field represents a Message's field
type Field struct {
	Offset      Offset
	Name        string
	Comments    []string
	Annotations AnnotationCollection
	Type        Type
	Index       int
}

// OneOfField represents an oneof field present in a Message
type OneOfField struct {
	Offset      Offset
	Comments    []string
	Annotations AnnotationCollection
	Index       int
	Items       []any
}

type parser struct {
	annotations AnnotationCollection
	comments    []string
	file        *File
	tokens      *tokenList
}

// Parse takes a list of Token and returns either a File, or an error.
func Parse(tokens []Token) (*File, error) {
	p := newParser(tokens)
	return p.run()
}

func newParser(tokens []Token) *parser {
	return &parser{
		annotations: nil,
		comments:    nil,
		file:        &File{},
		tokens: &tokenList{
			tokens:    tokens,
			tokensLen: len(tokens),
			current:   0,
		},
	}
}

func (p *parser) run() (*File, error) {
	if err := p.parsePackage(); err != nil {
		return nil, err
	}
	if err := p.parseImports(); err != nil {
		return nil, err
	}
	for !p.tokens.peek().is(EOF) {
		if err := p.parseOne(p.messageOrService); err != nil {
			return nil, err
		}
	}
	return p.file, nil
}

func (p *parser) messageOrService() error {
	if !p.tokens.peek().is(Identifier) {
		return p.tokens.error("expected identifier")
	}

	switch p.tokens.peek().Value {
	case "message":
		return p.message()
	case "service":
		return p.service()
	case "import":
		return p.tokens.error("imports are only allowed in the beginning of the file, after the package directive.")
	default:
		return p.tokens.error("unexpected `%s', expected 'message', 'service'", p.tokens.peek().Value)
	}
}

func (p *parser) message() error {
	start := p.tokens.advance() // consume "message"
	if !p.tokens.peek().is(Identifier) {
		return p.tokens.error("expected identifier")
	}
	name := p.tokens.peek()
	if p.file.isDefined(name.Value) {
		return p.tokens.error("%s is already defined", name.Value)
	}
	p.tokens.advance()
	if !p.tokens.peek().is(OpenCurly) {
		return p.tokens.error("expected '{'")
	}

	m := Message{
		Offset:      Offset{},
		Name:        name.Value,
		Comments:    p.comments,
		Annotations: p.annotations,
		Fields:      nil,
	}
	p.tokens.advance() // consume curly
	p.flushMeta()
	for !p.tokens.peek().is(CloseCurly) {
		err := p.parseOne(func() error {
			return p.parseStructureField(&m.Fields, true)
		})
		if err != nil {
			return err
		}
	}
	end := p.tokens.advance() // consume curly
	m.Offset = offsetBetween(start, end)
	p.file.push(m)
	return nil
}

func (p *parser) parseStructureField(arr *[]any, allowOneOf bool) error {
	if !p.tokens.peek().is(Identifier) {
		return p.tokens.error("expected identifier")
	}
	if p.tokens.peek().Value == "oneof" {
		if !allowOneOf {
			return p.tokens.error("oneof field is not allowed at this point")
		}
		return p.parseOneOf(arr)
	}

	fName := p.tokens.advance() // consume name
	fType, err := p.parseType()
	if err != nil {
		return err
	}
	fIndex, err := p.parseIndex()
	if err != nil {
		return err
	}
	if !p.tokens.peek().is(Semi) {
		return p.tokens.error("expected ';'")
	}
	end := p.tokens.advance()
	*arr = append(*arr, Field{
		Offset:      offsetBetween(fName, end),
		Name:        fName.Value,
		Comments:    p.comments,
		Annotations: p.annotations,
		Type:        fType,
		Index:       fIndex,
	})
	p.flushMeta()
	return nil
}

func (p *parser) parseOneOf(arr *[]any) error {
	start := p.tokens.advance()
	if !p.tokens.peek().is(OpenCurly) {
		return p.tokens.error("expected '{'")
	}
	p.tokens.advance() // consume curly
	var items []any
	comments := p.comments
	annotations := p.annotations
	p.flushMeta()
	for !p.tokens.peek().is(CloseCurly) {
		if err := p.parseOne(func() error {
			return p.parseStructureField(&items, false)
		}); err != nil {
			return err
		}
	}
	p.tokens.advance() // consume closeCurly
	idx, err := p.parseIndex()
	if err != nil {
		return err
	}
	if !p.tokens.peek().is(Semi) {
		return p.tokens.error("expected ';'")
	}
	end := p.tokens.advance()
	*arr = append(*arr, OneOfField{
		Offset:      offsetBetween(start, end),
		Comments:    comments,
		Annotations: annotations,
		Index:       idx,
		Items:       items,
	})
	return nil
}

func (p *parser) parseOne(or func() error) error {
	current := p.tokens.peek()
	switch current.Type {
	case LineBreak:
		if p.tokens.peekPrevious().is(LineBreak) {
			p.flushMeta()
		}
		p.tokens.advance()
	case Annotation:
		start := p.tokens.advance()
		end := start
		var vals []string
		if p.tokens.peek().is(OpenParen) {
			var val []string
			for !p.tokens.peek().is(CloseParen) {
				if p.tokens.peek().is(Comma) {
					if len(vals) == 0 {
						return p.tokens.error("expected value")
					}
					vals = append(vals, strings.Join(val, " "))
					val = val[:0]
					p.tokens.advance() // consume comma
					continue
				}
				val = append(val, p.tokens.advance().Value)
			}
			if len(val) > 0 {
				vals = append(vals, strings.Join(val, " "))
			}
			end = p.tokens.advance()
		}

		p.annotations = append(p.annotations, AnnotationValue{
			Offset: offsetBetween(start, end),
			Name:   start.Value,
			Value:  vals,
		})
	case Comment:
		push := p.tokens.peekPrevious().is(LineBreak)
		cmm := p.tokens.advance().Value
		if push {
			p.comments = append(p.comments, cmm)
		}
	default:
		return or()
	}

	return nil
}

func (p *parser) parsePackage() error {
	for p.tokens.peek().is(LineBreak) || p.tokens.peek().is(Comment) {
		p.tokens.advance()
	}
	if !p.tokens.peek().is(Identifier) {
		return p.tokens.error("expected identifier")
	}
	if p.tokens.peek().Value != "package" {
		return p.tokens.error("unexpected %s, expected package identifier", p.tokens.peek().Value)
	}
	start := p.tokens.advance() // consume package

	pName := []string{p.tokens.advance().Value}
	for p.tokens.peek().is(Identifier) || p.tokens.peek().is(Dot) {
		pName = append(pName, p.tokens.advance().Value)
	}
	if !p.tokens.peek().is(Semi) {
		return p.tokens.error("expected ';'")
	}
	end := p.tokens.advance()
	p.file.push(Package{
		Offset: offsetBetween(start, end),
		Name:   strings.Join(pName, ""),
	})

	return nil
}

func (p *parser) parseImports() error {
	for {
		for {
			if p.tokens.peek().is(LineBreak) {
				if p.tokens.peekPrevious().is(LineBreak) {
					p.flushMeta()
				}

				p.tokens.advance()
			} else if p.tokens.peek().is(Comment) {
				p.comments = append(p.comments, p.tokens.advance().Value)
			} else {
				break
			}
		}

		if !p.tokens.peek().is(Identifier) {
			return nil
		}

		if p.tokens.peek().Value != "import" {
			return nil
		}

		p.flushMeta()
		start := p.tokens.advance() // consume import

		if !p.tokens.peek().is(StringElement) {
			return p.tokens.error("expected string")
		}
		path := p.tokens.advance().Value //consume string
		if p.file.isImported(path) {
			return p.tokens.error("duplicated import")
		}
		if !p.tokens.peek().is(Semi) {
			return p.tokens.error("expected ';'")
		}
		end := p.tokens.advance() // consume semi
		p.file.push(Import{
			Offset: offsetBetween(start, end),
			Path:   path,
		})
	}
}

func (p *parser) annotation() error {
	start := p.tokens.advance() // annotation
	annot := AnnotationValue{
		Name: start.Value,
	}
	end := start
	if p.tokens.peek().Type == OpenCurly {
		var items []string
		var item []string
		for !p.tokens.peek().is(CloseCurly) {
			if p.tokens.peek().is(Comma) {
				items = append(items, strings.Join(item, " "))
				item = item[:0]
			} else {
				item = append(item, p.tokens.peek().Value)
			}
			p.tokens.advance()
		}
		if len(item) > 0 {
			items = append(items, strings.Join(item, " "))
		}
		end = p.tokens.advance()
		annot.Value = items
	}
	annot.Offset = offsetBetween(start, end)
	p.annotations = append(p.annotations, annot)
	return nil
}

func offsetBetween(a, b Token) Offset {
	return Offset{
		StartsAt: Position{
			Line:   a.Line,
			Column: a.Column,
		},
		EndsAt: Position{
			Line:   b.Line,
			Column: b.Column,
		},
	}
}

var stringToPrimitive = map[string]PrimitiveType{
	"string":  String,
	"uint8":   Uint8,
	"uint16":  Uint16,
	"uint32":  Uint32,
	"uint64":  Uint64,
	"int8":    Int8,
	"int16":   Int16,
	"int32":   Int32,
	"int64":   Int64,
	"float32": Float32,
	"float64": Float64,
	"bool":    Bool,
}

func (p *parser) parseType() (Type, error) {
	if !p.tokens.peek().is(Identifier) {
		return nil, p.tokens.error("unexpected token")
	}
	t := p.tokens.advance().Value

	if v, ok := stringToPrimitive[t]; ok {
		return Primitive{Kind: v}, nil
	}

	switch t {
	case "array":
		return p.parseArrayType()
	case "map":
		return p.parseMapType()
	default:
		v := []string{t}
		for p.tokens.peek().is(Identifier) || p.tokens.peek().is(Dot) {
			v = append(v, p.tokens.advance().Value)
		}
		return Unresolved{Name: strings.Join(v, "")}, nil
	}
}

func (p *parser) parseMapType() (Type, error) {
	k, err := p.parseMapKey()
	if err != nil {
		return nil, err
	}
	if !p.tokens.peek().is(Comma) {
		return nil, fmt.Errorf("expected ','")
	}
	p.tokens.advance()
	v, err := p.parseType()
	if err != nil {
		return nil, err
	}
	if !p.tokens.peek().is(CloseAngled) {
		return nil, fmt.Errorf("expected '>'")
	}
	p.tokens.advance()
	return Map{
		Key:   k,
		Value: v,
	}, nil
}

func (p *parser) parseMapKey() (PrimitiveType, error) {
	if !p.tokens.peek().is(Identifier) {
		return Invalid, p.tokens.error("unexpected token")
	}
	k := p.tokens.advance().Value
	v, ok := stringToPrimitive[k]
	if !ok || v == Bool {
		validKeys := make([]string, 0, len(stringToPrimitive))
		for k := range stringToPrimitive {
			validKeys = append(validKeys, k)
		}
		return Invalid, p.tokens.error("invalid type for map key, expected one of %s", strings.Join(validKeys, ", "))
	}
	return v, nil
}

func (p *parser) parseArrayType() (Type, error) {
	if !p.tokens.peek().is(OpenAngled) {
		return nil, p.tokens.error("expected '<")
	}
	p.tokens.advance()
	t, err := p.parseType()
	if err != nil {
		return nil, err
	}
	if !p.tokens.peek().is(CloseAngled) {
		return nil, p.tokens.error("expected '>")
	}
	p.tokens.advance()

	return Array{Of: t}, nil
}

func (p *parser) parseIndex() (int, error) {
	if !p.tokens.peek().is(Equal) {
		return 0, p.tokens.error("expected '='")
	}
	p.tokens.advance() // consume '='
	if !p.tokens.peek().is(Number) {
		return 0, p.tokens.error("expected number")
	}
	return strconv.Atoi(p.tokens.advance().Value)
}

func (p *parser) flushMeta() {
	p.comments = p.comments[:0]
	p.annotations = p.annotations[:0]
}

func (p *parser) service() error {
	start := p.tokens.advance() // consume "message"
	if !p.tokens.peek().is(Identifier) {
		return p.tokens.error("expected identifier")
	}

	name := p.tokens.peek()
	if p.file.isDefined(name.Value) {
		return p.tokens.error("%s is already defined", name.Value)
	}
	p.tokens.advance()

	if !p.tokens.peek().is(OpenCurly) {
		return p.tokens.error("expected '{'")
	}
	p.tokens.advance() // consume curly
	s := Service{
		Offset:      Offset{},
		Name:        name.Value,
		Comments:    p.comments,
		Annotations: p.annotations,
		Methods:     nil,
	}
	p.flushMeta()
	for !p.tokens.peek().is(CloseCurly) {
		if err := p.parseOne(p.parseMethod(&s)); err != nil {
			return err
		}
	}
	end := p.tokens.advance()
	s.Offset = offsetBetween(start, end)
	p.file.push(s)
	return nil
}

func (p *parser) parseMethod(s *Service) func() error {
	return func() error {
		if !p.tokens.peek().is(Identifier) {
			return p.tokens.error("expected identifier")
		}
		name := p.tokens.advance()
		if !p.tokens.peek().is(OpenParen) {
			return p.tokens.error("expected '('")
		}
		p.tokens.advance() // consume paren
		reqType := "void"
		if !p.tokens.peek().is(Identifier) && !p.tokens.peek().is(CloseParen) {
			return p.tokens.error("expected identifier or ')'")
		}

		for p.tokens.peek().is(Identifier) || p.tokens.peek().is(Dot) {
			if reqType == "void" {
				reqType = p.tokens.advance().Value
			} else {
				reqType += p.tokens.advance().Value
			}
		}

		if !p.tokens.peek().is(CloseParen) {
			return p.tokens.error("expected ')'")
		}
		p.tokens.advance() // consume paren
		retType := "void"
		stream := false
		if !p.tokens.peek().is(Semi) {
			retType = ""
			if !p.tokens.peek().is(Arrow) {
				return p.tokens.error("expected '->'")
			}
			p.tokens.advance() // consume arrow
			if p.tokens.peek().is(Identifier) && p.tokens.peek().Value == "stream" {
				p.tokens.advance() // consume stream
				stream = true
			}

			if !p.tokens.peek().is(Identifier) {
				return p.tokens.error("expected identifier")
			}
			for p.tokens.peek().is(Identifier) || p.tokens.peek().is(Dot) {
				retType += p.tokens.advance().Value
			}
		}

		if !p.tokens.peek().is(Semi) {
			return p.tokens.error("expected ';'")
		}
		end := p.tokens.advance()
		s.Methods = append(s.Methods, Method{
			Offset:          offsetBetween(name, end),
			Name:            name.Value,
			Comments:        p.comments,
			Annotations:     p.annotations,
			ArgumentType:    reqType,
			ReturnType:      retType,
			ReturnStreaming: stream,
		})
		p.flushMeta()
		return nil
	}
}
