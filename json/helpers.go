package json

func isDecDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') ||
	    (r >= 'a' && r <= 'f') ||
	    (r >= 'A' && r <= 'F')
}
