package idl

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

// Scanner implements mechanisms responsible for reading an IDL file into a list
// of tokens.
type Scanner struct {
	tokens  []Token
	data    []rune
	dataLen int
	start   int
	current int
}

// Scan takes an io.Reader and returns a list of Token from it, or an error, in
// case the file is invalid. This is a convenience function that creates a new
// Scanner, reads the provided io.Reader into it, and returns the resulting
// value. Scan does not close the provided io.Reader.
func Scan(r io.Reader) ([]Token, error) {
	s, err := NewScanner(r)
	if err != nil {
		return nil, err
	}
	return s.Run()
}

// NewScanner creates a new Scanner bound to a given io.Reader. The scanner does
// not close the provided reader.
// See also: Scan
func NewScanner(r io.Reader) (*Scanner, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &Scanner{
		tokens:  nil,
		data:    []rune(string(buf)),
		dataLen: len(buf),
		start:   0,
		current: 0,
	}, nil
}

// Run executes the scan process into the provided reader. Returns either a list
// of Token, or an error.
func (s *Scanner) Run() ([]Token, error) {
	for !s.isAtEnd() {
		s.start = s.current
		if err := s.scanToken(); err != nil {
			return nil, err
		}
	}
	s.pushToken(EOF, "")
	return s.tokens, nil
}

func (s *Scanner) pushToken(k Element, v string) {
	l, c := s.pos()
	s.tokens = append(s.tokens, Token{
		Type:   k,
		Value:  v,
		Line:   l,
		Column: c,
	})
}

var simpleTokens = map[rune]Element{
	'(':  OpenParen,
	')':  CloseParen,
	'<':  OpenAngled,
	'>':  CloseAngled,
	'{':  OpenCurly,
	'}':  CloseCurly,
	',':  Comma,
	'.':  Dot,
	'=':  Equal,
	';':  Semi,
	'\n': LineBreak,
}

func (s *Scanner) scanToken() error {
	r := s.advance()
	switch r {
	case '@':
		if err := s.annotation(); err != nil {
			return err
		}
	case '-':
		if s.peek() != '>' {
			unkChar := s.advance()
			return s.error("Unexpected `%c', expected `>'", unkChar)
		}
		s.pushToken(Arrow, "->")
		// We advance later here so we can point the arrow to
		// the beginning of it instead of the end.
		s.advance()

	case '\r', ' ', '\t':
	// Just consume it. We don't care about spaces
	case '"':
		return s.string()
	case '#':
		s.comment()
	default:
		if k, ok := simpleTokens[r]; ok {
			s.pushToken(k, string(r))
		} else if unicode.IsDigit(r) {
			s.number()
		} else if unicode.IsGraphic(r) {
			s.identifier()
		} else {
			return s.error("Unexpected `%c'", r)
		}
	}

	return nil
}

func (s *Scanner) advance() rune {
	r := s.data[s.current]
	s.current++
	return r
}

func (s Scanner) peek() rune {
	if s.isAtEnd() {
		return 0x00
	}
	return s.data[s.current]
}

func (s Scanner) peekNext() rune {
	if s.current+1 >= s.dataLen {
		return 0x00
	}
	return s.data[s.current+1]
}

func (s Scanner) pos() (int, int) {
	line := 1
	column := 1
	for i := 0; i < s.current; i++ {
		if s.data[i] == '\n' {
			line++
			column = 1
		}
		column++
	}
	return line, column
}

func (s Scanner) isAtEnd() bool {
	return s.current >= len(s.data)
}

func (s Scanner) error(msg string, a ...interface{}) error {
	l, c := s.pos()
	return SyntaxError{
		Message: fmt.Sprintf(msg, a...),
		Line:    l,
		Column:  c,
	}
}

func (s *Scanner) number() {
	l, c := s.pos()
	for unicode.IsDigit(s.peek()) {
		s.advance()
	}
	s.tokens = append(s.tokens, Token{
		Type:   Number,
		Value:  string(s.data[s.start:s.current]),
		Line:   l,
		Column: c,
	})
}

func (s *Scanner) identifier() {
	l, col := s.pos()
	c := s.peek()
	if (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c == '_') {
		s.advance()
	}
	c = s.peek()
	for (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c == '_') ||
		(c >= '0' && c <= '9') {
		s.advance()
		c = s.peek()
	}

	s.tokens = append(s.tokens, Token{
		Type:   Identifier,
		Value:  string(s.data[s.start:s.current]),
		Line:   l,
		Column: col,
	})
}

func (s *Scanner) comment() {
	l, c := s.pos()
	for s.peek() != '\n' {
		s.advance()
	}
	s.tokens = append(s.tokens, Token{
		Type:   Comment,
		Value:  strings.TrimSpace(string(s.data[s.start+1 : s.current])),
		Line:   l,
		Column: c,
	})
}

func (s *Scanner) annotation() error {
	l, c := s.pos()
	for s.peek() != ' ' {
		s.advance()
	}
	consumed := s.current - s.start
	if consumed == 1 {
		return s.error("Unexpected `%c', expected identifier", s.peek())
	}
	s.tokens = append(s.tokens, Token{
		Type:   Annotation,
		Value:  string(s.data[s.start+1 : s.current]),
		Line:   l,
		Column: c,
	})
	return nil
}

func (s *Scanner) string() error {
	l, c := s.pos()
	s.advance() // consume "
	escaping := false

loop:
	for {
		if s.isAtEnd() {
			return s.error("unterminated string")
		}
		switch s.peek() {
		case '"':
			if !escaping {
				break loop
			}
			escaping = false
		case '\\':
			escaping = true
		case '\n':
			return s.error("unterminated string")
		}
		s.advance()
	}
	s.advance() // consume "

	s.tokens = append(s.tokens, Token{
		Type:   StringElement,
		Value:  strings.ReplaceAll(string(s.data[s.start+1:s.current-1]), "\\\"", `"`),
		Line:   l,
		Column: c,
	})

	return nil
}
