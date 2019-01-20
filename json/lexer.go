package json

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type (
	lexer struct {
		reader reader
		err    error // Optional error occurred during the parse
		// Represents a token that has been read the last time
		tok struct {
			kind  tokenKind
			pos   pos
			value string // Optional value of the token
		}
	}
)

const (
	// A set of characters, which must be ignored unless they
	// are a part of a string.

	ws  = '\x20' // whitespace
	tab = '\x09' // tab
	lf  = '\x0A' // line feed
	cr  = '\x0D' // carriage return

	// A default representation of a newline, which doesn't depend on how
	// the newlines are represented in the input string.
	newline = lf

	// A rune, which doesn't represent any specific value.
	eof = rune(0)
)

// -----------------------------------------------------------------------------
// Lexer implementation

// eof gets a value indicating whether an end of file has been reached.
func (lex *lexer) eof() bool { return lex.tok.kind == tkEof }

func (lex *lexer) next() {
	var err error
	var cur rune
	for {
		cur, err = lex.reader.peek()
		if err != nil {
			lex.err = err
			return
		}
		if cur == eof {
			lex.tok.kind = tkEof
			return
		}
		if cur != ws && cur != tab && cur != newline {
			break
		}
		lex.reader.accept()
	}
	lex.tok.pos = lex.reader.pos
	switch cur {
	case '{':
		lex.acceptTok(tkCrBrOpen)
	case '}':
		lex.acceptTok(tkCrBrClose)
	case '[':
		lex.acceptTok(tkSqBrOpen)
	case ']':
		lex.acceptTok(tkSqBrClose)
	case ',':
		lex.acceptTok(tkComma)
	case ':':
		lex.acceptTok(tkColon)
	default:
		var parsed bool
		if parsed, err = lex.tryParseString(); parsed {
			// Do nothing
		} else if parsed, err = lex.tryParseKeyword(); parsed {
			// Do nothing
		} else if parsed, err = lex.tryParseNumber(); parsed {
			// Do nothing
		} else {
			err = fmt.Errorf("dymessage: unexpected character '%c'", cur)
		}
	}
	lex.err = err
	return
}

func (lex *lexer) acceptTok(tk tokenKind) {
	lex.tok.kind = tk
	lex.reader.accept()
}

func (lex *lexer) tryParseString() (parsed bool, err error) {
	var buf strings.Builder
	var r rune
	if r, err = lex.reader.peekNoEof(); err != nil {
		return
	}
	if r != '"' {
		return
	}
	lex.reader.accept()
	parsed = true
IterateString:
	for {
		if r, err = lex.reader.peekNoEof(); err != nil {
			return
		}
		lex.reader.accept()
		switch r {
		case '\\':
			if r, err = lex.reader.peekNoEof(); err != nil {
				return
			}
			switch r {
			case '"', '\\', '/':
				lex.reader.accept()
				// r already contains the proper value
			case 'b':
				lex.reader.accept()
				r = '\x08' // backspace
			case 'f':
				lex.reader.accept()
				r = '\x0C' // form feed
			case 'n':
				lex.reader.accept()
				r = '\x0A' // line feed
			case 'r':
				lex.reader.accept()
				r = '\x0D' // carriage return
			case 't':
				lex.reader.accept()
				r = '\x09' // tab
			case 'u':
				lex.reader.accept()
				var uc [4]rune
				for i := 0; i < len(uc); i++ {
					if uc[i], err = lex.reader.peekHexNoEof(); err != nil {
						return
					}
					lex.reader.accept()
				}
				var n uint64
				n, err = strconv.ParseUint(string(uc[:]), 16, 32)
				if err != nil {
					return
				}
				r = rune(n)
			default:
				err = fmt.Errorf("dymessage: bad escape character '%c'", r)
				break
			}
		case '"':
			break IterateString
		}
		buf.WriteRune(r)
	}
	lex.tok.kind = tkString
	lex.tok.value = buf.String()
	return
}

func (lex *lexer) tryParseNumber() (parsed bool, err error) {
	var buf strings.Builder
	var r rune
	if r, err = lex.reader.peekNoEof(); err != nil {
		goto Error
	}
	if r == '-' || r == '+' {
		buf.WriteRune(r)
		lex.reader.accept()
		if r, err = lex.reader.peekDecNoEof(); err != nil {
			goto Error
		}
	}
	if r == '0' {
		buf.WriteRune(r)
		lex.reader.accept()
		if r, err = lex.reader.peek(); err != nil {
			goto Error
		}
		if isDecDigit(r) {
			err = fmt.Errorf("dymessage: unexpected digit after a leading zero")
			goto Error
		}
	} else {
		if !isDecDigit(r) {
			err = fmt.Errorf("dymessage: '%c' is not a valid decimal digit", r)
			goto Error
		}
		for {
			buf.WriteRune(r)
			lex.reader.accept()
			if r, err = lex.reader.peek(); err != nil {
				goto Error
			}
			if !isDecDigit(r) {
				break
			}
		}
	}
	if r == '.' {
		buf.WriteRune(r)
		lex.reader.accept()
		if r, err = lex.reader.peekDecNoEof(); err != nil {
			goto Error
		}
		for {
			buf.WriteRune(r)
			lex.reader.accept()
			if r, err = lex.reader.peek(); err != nil {
				goto Error
			}
			if !isDecDigit(r) {
				break
			}
		}
	}
	if r == 'e' || r == 'E' {
		buf.WriteRune(r)
		lex.reader.accept()
		if r, err = lex.reader.peekNoEof(); err != nil {
			goto Error
		}
		if r == '-' || r == '+' {
			buf.WriteRune(r)
			lex.reader.accept()
			if r, err = lex.reader.peekDecNoEof(); err != nil {
				goto Error
			}
		} else if !isDecDigit(r) {
			err = fmt.Errorf("dymessage: '%c' is not a valid decimal digit", r)
			goto Error
		}
		for {
			buf.WriteRune(r)
			lex.reader.accept()
			if r, err = lex.reader.peek(); err != nil {
				goto Error
			}
			if !isDecDigit(r) {
				break
			}
		}
	}
Error:
	parsed = buf.Len() > 0
	if parsed && err == nil {
		lex.tok.kind = tkNumber
		lex.tok.value = buf.String()
	}
	return
}

func (lex *lexer) tryParseKeyword() (parsed bool, err error) {
	var buf strings.Builder
	var r rune
	for {
		r, err = lex.reader.peek()
		if err != nil {
			break
		}
		if !unicode.IsLetter(r) {
			break
		}
		buf.WriteRune(r)
		lex.reader.accept()
	}
	parsed = buf.Len() > 0
	if parsed && err == nil {
		lex.tok.value = buf.String()
		switch lex.tok.value {
		case "true":
			lex.tok.kind = tkTrue
		case "false":
			lex.tok.kind = tkFalse
		case "null":
			lex.tok.kind = tkNull
		default:
			err = fmt.Errorf("dymessage: value '%s' is not a valid keyword", lex.tok.value)
		}
	}
	return
}
