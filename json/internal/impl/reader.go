package impl

type reader struct {
	buf []byte
	off int

	cur rune  // The character read the last time and now ready to be consumed
	pos Pos   // A zero-based position of the current rune
}

const (
	// A set of characters, which must be ignored unless they
	// are a part of a string.

	ws  = '\x20' // whitespace
	tab = '\x09' // tab
	lf  = '\x0A' // line feed
	cr  = '\x0D' // carriage return

	// A default representation of a newline, which doesn't depend on how
	// the newlines are represented in the input string.
	nl = lf

	// A rune, which represents an end of file.
	eof = rune(0)
)

// -----------------------------------------------------------------------------
// Reader implementation

func (rd *reader) reset(buf []byte) {
	rd.buf, rd.off = buf, 0
	rd.pos = Pos{line: 0, col: -1}
	rd.accept()
}

func (rd *reader) peek() rune { return rd.cur }

func (rd *reader) accept() {
	r := rd.acceptRune()
	if r == cr || r == lf {
		rd.pos.line++
		rd.pos.col = 0
		// Checking if the newline character has another complementary
		// character 0x0A or 0x0D, and if yes, skipping it.
		_r := rd.peekRune()
		if r != _r && (_r == cr || _r == lf) {
			rd.acceptRune()
		}
		r = nl
	} else {
		rd.pos.col++
	}
	rd.cur = r
}
