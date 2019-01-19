package json

type tokenKind int

const (
	tkString tokenKind = iota
	tkNumber
	tkBool
	tkCrBrOpen
	tkCrBrClose
	tkSqBrOpen
	tkSqBrClose
	tkColon
	tkComma
	tkNull
)

func (tk tokenKind) String() string {
	switch tk {
	case tkString:
		return "string"
	case tkNumber:
		return "number"
	case tkBool:
		return "boolean"
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
		panic(tk)
	}
}
