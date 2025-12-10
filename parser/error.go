package parser

import (
	"fmt"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/token"
)

const (
	errUnexpectedToken      = "Unexpected token %v"
	errUnexpectedEndOfInput = "Unexpected end of input"
)

// SyntaxError represents a parsing error with position information
type SyntaxError struct {
	Message string
	Line    int
	Column  int
	Offset  int
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("%s (line %d, column %d)", e.Message, e.Line, e.Column)
}

// positionToLineColumn converts a byte offset to line and column numbers
func positionToLineColumn(src string, offset int) (line, col int) {
	line = 1
	col = 1
	for i := 0; i < offset && i < len(src); i++ {
		if src[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

// error ...
func (p *parser) error(msg string, msgValues ...any) error {
	msg = fmt.Sprintf(msg, msgValues...)
	p.errors.Add(p.str, p.idx, msg)
	return p.errors[len(p.errors)-1]
}

// errorUnexpected ...
func (p *parser) errorUnexpected(chr rune) error {
	if chr == -1 {
		return p.error(errUnexpectedEndOfInput)
	}
	return p.error(errUnexpectedToken, token.Illegal)
}

func (p *parser) errorUnexpectedToken(tkn token.Token) error {
	switch tkn {
	case token.Eof:
		return p.error(errUnexpectedEndOfInput)
	}
	value := tkn.String()
	switch tkn {
	case token.Boolean, token.Null:
		value = p.literal
	case token.Identifier:
		return p.error("Unexpected identifier")
	case token.Keyword:
		// TODO Might be a future reserved word
		return p.error("Unexpected reserved word")
	case token.EscapedReservedWord:
		return p.error("Keyword must not contain escaped characters")
	case token.Number:
		return p.error("Unexpected number")
	case token.String:
		return p.error("Unexpected string")
	}
	return p.error(errUnexpectedToken, value)
}

// ErrorList is a list of *Errors.
type ErrorList []*SyntaxError

// Add adds an Error with given position and message to an ErrorList.
func (e *ErrorList) Add(src string, idx ast.Idx, msg string) {
	offset := int(idx) - 1 // Convert 1-based idx to 0-based offset
	if offset < 0 {
		offset = 0
	}
	line, col := positionToLineColumn(src, offset)
	*e = append(*e, &SyntaxError{
		Message: msg,
		Line:    line,
		Column:  col,
		Offset:  offset,
	})
}

// Error implements the Error interface.
func (e *ErrorList) Error() string {
	switch len(*e) {
	case 0:
		return "no errors"
	case 1:
		return (*e)[0].Error()
	}
	return fmt.Sprintf("%s (and %d more errors)", (*e)[0].Error(), len(*e)-1)
}

// Err returns an error equivalent to this ErrorList. If the list is empty, Err returns nil.
func (e *ErrorList) Err() error {
	if len(*e) == 0 {
		return nil
	}
	return e
}
