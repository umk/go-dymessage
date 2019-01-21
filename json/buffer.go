package json

import (
	"io"
	"unicode/utf8"
)

type buffer struct {
	buf []byte
	off int
}

func (buf *buffer) peekRune() (r rune, err error) {
	r, _, err = buf.peekRuneImpl()
	return
}

func (buf *buffer) readRune() (r rune, err error) {
	var size int
	if r, size, err = buf.peekRuneImpl(); err == nil {
		buf.off += size
	}
	return
}

func (buf *buffer) peekRuneImpl() (rune, int, error) {
	if buf.off == len(buf.buf) {
		return rune(0), 0, io.EOF
	}
	r, size := rune(buf.buf[buf.off]), 1
	if r >= utf8.RuneSelf {
		r, size = utf8.DecodeRune(buf.buf[buf.off:])
	}
	return r, size, nil
}
