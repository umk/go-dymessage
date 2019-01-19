package json

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type (
	reader struct {
		rd *bufio.Reader

		cur rune  // The character read the last time ready to be consumed
		pos pos   // A zero-based position of the current rune
		err error // An error occurred when reading the latest character
	}

	// Represents a zero-based position of the rune in the text.
	pos struct{ line, col int }

	lexer struct {
		reader reader

		// Represents a token that has been read the last time
		tok struct {
			kind tokenKind
			pos  pos

			// The following are the values, which may belong to current token.

			string string
			number string
			bool   bool
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
	none = rune(0)
)

// The literals, which are a part of the JSON syntax.
const (
	litTrue  = "true"
	litFalse = "false"
	litNull  = "null"
)

// -----------------------------------------------------------------------------
// Reader implementation

func (rd *reader) reset(r io.Reader) {
	rd.rd, rd.pos = bufio.NewReader(r), pos{line: 0, col: -1}
	rd.accept()
}

func (rd *reader) peek() (rune, error) {
	return rd.cur, rd.err
}

func (rd *reader) accept() {
	r, _, err := rd.rd.ReadRune()
	if err != nil {
		if err == io.EOF {
			// The end of file must be recognized by the none rune.
			err = nil
		}
		rd.cur, rd.err = none, err
		return
	}
	if r == cr || r == lf {
		// Checking if the newline character has another complementary
		// character 0x0A or 0x0D, and if yes, skipping it.
		_r, _, _err := rd.rd.ReadRune()
		if _err != nil {
			_ = rd.rd.UnreadRune()
		} else if r == _r || (_r != cr && _r != lf) {
			_ = rd.rd.UnreadRune()
		}
		rd.pos.line++
		rd.pos.col = 0
		r = newline
	} else {
		rd.pos.col++
	}
	rd.cur, rd.err = r, nil
}

// -----------------------------------------------------------------------------
// Special peek methods

func (rd *reader) peekNoEof() (r rune, err error) {
	r, err = rd.peek()
	if err == nil && r == none {
		err = io.ErrUnexpectedEOF
	}
	return
}

func (rd *reader) peekDec() (r rune, err error) {
	r, err = rd.peek()
	if r != none && err == nil && !isDecDigit(r) {
		err = fmt.Errorf("dymessage: '%c' is not a valid decimal digit", r)
	}
	return
}

func (rd *reader) peekDecNoEof() (r rune, err error) {
	r, err = rd.peekDec()
	if r == none {
		err = io.ErrUnexpectedEOF
	}
	return
}

func (rd *reader) peekHexNoEof() (r rune, err error) {
	r, err = rd.peekNoEof()
	if err == nil && !isHexDigit(r) {
		err = fmt.Errorf("dymessage: '%c' is not a valid hex digit", r)
	}
	return
}

// -----------------------------------------------------------------------------
// Lexer implementation

func (lex *lexer) next() (err error) {
	var cur rune
	for {
		cur, err = lex.reader.peek()
		if err != nil {
			return
		}
		if cur == none {
			return io.EOF
		}
		if cur != ws && cur != tab && cur != newline {
			break
		}
		lex.reader.accept()
	}
	lex.tok.pos = lex.reader.pos
	switch cur {
	case '{':
		lex.consumeTok(tkCrBrOpen)
	case '}':
		lex.consumeTok(tkCrBrClose)
	case '[':
		lex.consumeTok(tkSqBrOpen)
	case ']':
		lex.consumeTok(tkSqBrClose)
	case ',':
		lex.consumeTok(tkComma)
	case ':':
		lex.consumeTok(tkColon)
	default:
		var handled bool
		if handled, err = lex.handleString(); handled {
			// Do nothing
		} else if handled, err = lex.handleKeyword(); handled {
			// Do nothing
		} else if handled, err = lex.handleNumber(); handled {
			// Do nothing
		} else {
			err = fmt.Errorf("dymessage: unexpected character '%c'", cur)
		}
	}
	return
}

func (lex *lexer) consumeTok(tk tokenKind) {
	lex.tok.kind = tk
	lex.reader.accept()
}

func (lex *lexer) handleString() (handled bool, err error) {
	var buf strings.Builder
	var r rune
	if r, err = lex.reader.peekNoEof(); err != nil {
		return
	}
	if r != '"' {
		return
	}
	lex.reader.accept()
	handled = true
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
	lex.tok.string = buf.String()
	return
}

func (lex *lexer) handleNumber() (handled bool, err error) {
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
	handled = buf.Len() > 0
	if handled && err == nil {
		lex.tok.kind = tkNumber
		lex.tok.number = buf.String()
	}
	return
}

func (lex *lexer) handleKeyword() (handled bool, err error) {
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
	handled = buf.Len() > 0
	if handled && err == nil {
		keyword := buf.String()
		switch keyword {
		case litTrue:
			lex.tok.kind, lex.tok.bool = tkBool, true
		case litFalse:
			lex.tok.kind, lex.tok.bool = tkBool, false
		case litNull:
			lex.tok.kind = tkNull
		default:
			err = fmt.Errorf("dymessage: value '%s' is not a valid keyword", keyword)
		}
	}
	return
}
