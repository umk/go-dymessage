package impl

import (
	"fmt"
	"io"
)

func (rd *reader) peekNoEof() (r rune, err error) {
	r, err = rd.peek()
	if err == nil && r == eof {
		err = io.ErrUnexpectedEOF
	}
	return
}

func (rd *reader) peekDec() (r rune, err error) {
	r, err = rd.peek()
	if r != eof && err == nil && !isDecDigit(r) {
		err = fmt.Errorf("dymessage: '%c' is not a valid decimal digit", r)
	}
	return
}

func (rd *reader) peekDecNoEof() (r rune, err error) {
	r, err = rd.peekDec()
	if r == eof {
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
