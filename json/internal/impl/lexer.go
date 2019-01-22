package impl

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type Lexer struct {
	reader reader
	Err    error // Optional error occurred during the parse
	// Represents a token that has been read the last time
	Tok struct {
		Kind  TokenKind
		Pos   Pos // A zero-based line and column indexes of the token
		Value string // Optional value of the token
	}
}

// -----------------------------------------------------------------------------
// Lexer implementation

// Reset prepares the lexer for new deserialization by assigning it with a
// buffer that contains the input JSON.
func (lex *Lexer) Reset(buf []byte) { lex.reader.reset(buf) }

// Eof gets a value indicating whether an end of file has been reached.
func (lex *Lexer) Eof() bool { return lex.Tok.Kind == TkEof }

func (lex *Lexer) Next() {
	var err error
	var cur rune
	for {
		if cur = lex.reader.peek(); cur == eof {
			lex.Tok.Kind = TkEof
			return
		}
		if cur != ws && cur != tab && cur != nl {
			break
		}
		lex.reader.accept()
	}
	lex.Tok.Pos = lex.reader.pos
	switch cur {
	case '{':
		lex.acceptTok(TkCrBrOpen)
	case '}':
		lex.acceptTok(TkCrBrClose)
	case '[':
		lex.acceptTok(TkSqBrOpen)
	case ']':
		lex.acceptTok(TkSqBrClose)
	case ',':
		lex.acceptTok(TkComma)
	case ':':
		lex.acceptTok(TkColon)
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
	lex.Err = err
	return
}

func (lex *Lexer) acceptTok(tk TokenKind) {
	lex.Tok.Kind = tk
	lex.reader.accept()
}

func (lex *Lexer) tryParseString() (parsed bool, err error) {
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
	lex.Tok.Kind = TkString
	lex.Tok.Value = buf.String()
	return
}

func (lex *Lexer) tryParseNumber() (parsed bool, err error) {
	var buf strings.Builder
	var r rune
	if r, err = lex.reader.peekNoEof(); err != nil {
		goto DoneOrError
	}
	if r == '-' || r == '+' {
		buf.WriteRune(r)
		lex.reader.accept()
		if r, err = lex.reader.peekDecNoEof(); err != nil {
			goto DoneOrError
		}
	}
	if r == '0' {
		buf.WriteRune(r)
		lex.reader.accept()
		if r = lex.reader.peek(); r == eof {
			goto DoneOrError
		}
		if isDecDigit(r) {
			err = fmt.Errorf("dymessage: unexpected digit after a leading zero")
			goto DoneOrError
		}
	} else {
		if !isDecDigit(r) {
			err = fmt.Errorf("dymessage: '%c' is not a valid decimal digit", r)
			goto DoneOrError
		}
		for {
			buf.WriteRune(r)
			lex.reader.accept()
			if r = lex.reader.peek(); r == eof {
				goto DoneOrError
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
			goto DoneOrError
		}
		for {
			buf.WriteRune(r)
			lex.reader.accept()
			if r = lex.reader.peek(); r == eof {
				goto DoneOrError
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
			goto DoneOrError
		}
		if r == '-' || r == '+' {
			buf.WriteRune(r)
			lex.reader.accept()
			if r, err = lex.reader.peekDecNoEof(); err != nil {
				goto DoneOrError
			}
		} else if !isDecDigit(r) {
			err = fmt.Errorf("dymessage: '%c' is not a valid decimal digit", r)
			goto DoneOrError
		}
		for {
			buf.WriteRune(r)
			lex.reader.accept()
			if r = lex.reader.peek(); r == eof {
				goto DoneOrError
			}
			if !isDecDigit(r) {
				break
			}
		}
	}
DoneOrError:
	parsed = buf.Len() > 0
	if parsed && err == nil {
		lex.Tok.Kind = TkNumber
		lex.Tok.Value = buf.String()
	}
	return
}

func (lex *Lexer) tryParseKeyword() (parsed bool, err error) {
	var buf strings.Builder
	var r rune
	for {
		if r = lex.reader.peek(); r == eof {
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
		lex.Tok.Value = buf.String()
		switch lex.Tok.Value {
		case "true":
			lex.Tok.Kind = TkTrue
		case "false":
			lex.Tok.Kind = TkFalse
		case "null":
			lex.Tok.Kind = TkNull
		default:
			err = fmt.Errorf("dymessage: value '%s' is not a valid keyword", lex.Tok.Value)
		}
	}
	return
}
