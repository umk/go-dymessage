package impl

import "fmt"

type (
	TokenKind int

	// Represents a zero-based position of the token in the input.
	Pos struct{ line, col int }
)

const (
	TkEof TokenKind = iota
	TkString
	TkNumber
	TkTrue
	TkFalse
	TkCrBrOpen
	TkCrBrClose
	TkSqBrOpen
	TkSqBrClose
	TkColon
	TkComma
	TkNull
)

// MessageString gets a representation of the token kind based on whether it
// corresponds to a known character sequence or must correspond an arbitrary
// value.
func (tk TokenKind) MessageString() string {
	switch tk {
	case TkEof, TkString, TkNumber:
		return tk.String()
	default:
		return fmt.Sprintf("%q", tk)
	}
}

func (tk TokenKind) String() string {
	switch tk {
	case TkEof:
		return "EOF"
	case TkString:
		return "string"
	case TkNumber:
		return "number"
	case TkTrue:
		return "true"
	case TkFalse:
		return "false"
	case TkCrBrOpen:
		return "{"
	case TkCrBrClose:
		return "}"
	case TkSqBrOpen:
		return "["
	case TkSqBrClose:
		return "]"
	case TkColon:
		return ":"
	case TkComma:
		return ","
	case TkNull:
		return "null"
	default:
		panic(fmt.Sprintf("dymessage: invalid token kind %d", tk))
	}
}

func (pos Pos) String() string {
	return fmt.Sprintf("(%d:%d)", pos.line+1, pos.col+1)
}
