package json

import (
	"io"
	"unicode/utf8"
)

func (rd *reader) peekRune() (r rune, err error) {
	if rd.off == len(rd.buf) {
		err = io.EOF
	} else {
		r = rune(rd.buf[rd.off])
		if r >= utf8.RuneSelf {
			r, _ = utf8.DecodeRune(rd.buf[rd.off:])
		}
	}
	return
}

func (rd *reader) acceptRune() (r rune, err error) {
	if rd.off == len(rd.buf) {
		return r, io.EOF
	} else {
		var size int
		r, size = rune(rd.buf[rd.off]), 1
		if r >= utf8.RuneSelf {
			r, size = utf8.DecodeRune(rd.buf[rd.off:])
		}
		rd.off += size
		return
	}
}
