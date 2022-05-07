package idl

import "fmt"

type tokenList struct {
	tokens    []Token
	tokensLen int
	current   int
}

func (t tokenList) peek() Token {
	if t.current >= t.tokensLen {
		return Token{Type: EOF}
	}

	return t.tokens[t.current]
}

func (t tokenList) peekNext() Token {
	if t.current+1 >= t.tokensLen {
		return Token{Type: EOF}
	}

	return t.tokens[t.current+1]
}

func (t tokenList) peekPrevious() Token {
	if t.current == 0 {
		return t.tokens[t.current]
	}
	return t.tokens[t.current-1]
}

func (t *tokenList) advance() Token {
	current := t.peek()
	t.current++
	return current
}

func (t tokenList) error(msg string, a ...any) error {
	return ParseError{
		Token:   t.peek(),
		Message: fmt.Sprintf(msg, a...),
	}
}

func (t *tokenList) matchOrFail(el Element) error {
	if t.peek().is(el) {
		t.advance()
		return nil
	}
	return t.error("expected %s", el)
}
