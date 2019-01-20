package json

import "fmt"

type (
	tokenKind int

	// Represents a zero-based position of the token in the input.
	pos struct{ line, col int }
)

const (
	tkEof tokenKind = iota
	tkString
	tkNumber
	tkTrue
	tkFalse
	tkCrBrOpen
	tkCrBrClose
	tkSqBrOpen
	tkSqBrClose
	tkColon
	tkComma
	tkNull
)

// messageString gets a representation of the token kind based on whether it is
// represented by a known character sequence or can match
func (tk tokenKind) messageString() string {
	switch tk {
	case tkEof, tkString, tkNumber:
		return tk.String()
	default:
		return fmt.Sprintf("%q", tk)
	}
}

func (tk tokenKind) String() string {
	switch tk {
	case tkEof:
		return "eof"
	case tkString:
		return "string"
	case tkNumber:
		return "number"
	case tkTrue:
		return "true"
	case tkFalse:
		return "false"
	case tkCrBrOpen:
		return "{"
	case tkCrBrClose:
		return "}"
	case tkSqBrOpen:
		return "["
	case tkSqBrClose:
		return "]"
	case tkColon:
		return ":"
	case tkComma:
		return ","
	case tkNull:
		return "null"
	default:
		panic(fmt.Sprintf("dymessage: invalid token kind %d", tk))
	}
}

func (pos pos) String() string {
	return fmt.Sprintf("(%d:%d)", pos.line+1, pos.col+1)
}
