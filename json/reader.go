package json

import (
	"fmt"
	"io"
)

type reader struct {
	buf []byte
	off int

	cur rune  // The character read the last time ready to be consumed
	pos pos   // A zero-based position of the current rune
	err error // An error occurred when reading the latest character
}

// -----------------------------------------------------------------------------
// Reader implementation

func (rd *reader) reset(buf []byte) {
	rd.buf, rd.off = buf, 0
	rd.pos = pos{line: 0, col: -1}
	rd.accept()
}

func (rd *reader) peek() (rune, error) {
	return rd.cur, rd.err
}

func (rd *reader) accept() {
	r, err := rd.acceptRune()
	if err != nil {
		if err == io.EOF {
			// The end of file must be recognized by the eof rune.
			rd.cur, rd.err = eof, nil
		} else {
			rd.err = err
		}
		return
	}
	if rd.cur == newline {
		rd.pos.line++
		rd.pos.col = 0
	} else {
		rd.pos.col++
	}
	if r == cr || r == lf {
		// Checking if the newline character has another complementary
		// character 0x0A or 0x0D, and if yes, skipping it.
		_r, _err := rd.peekRune()
		if _err != nil {
			//
		} else if r == _r || (_r != cr && _r != lf) {
			//
		} else {
			_, _ = rd.acceptRune()
		}
		r = newline
	}
	rd.cur, rd.err = r, nil
}

// -----------------------------------------------------------------------------
// Ad hoc peek methods

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
