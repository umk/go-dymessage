package json

import (
	"bufio"
	"fmt"
	"io"
)

type reader struct {
	rd *bufio.Reader

	cur rune  // The character read the last time ready to be consumed
	pos pos   // A zero-based position of the current rune
	err error // An error occurred when reading the latest character
}

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
			// The end of file must be recognized by the eof rune.
			rd.cur, rd.err = eof, nil
		} else {
			rd.err = err
		}
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
