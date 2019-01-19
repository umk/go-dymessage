package json

import (
	"encoding/json"
	"fmt"
)

type token struct {
	tok interface{}
	err error
}

func (tok token) delim(r rune) error {
	if tok.err != nil {
		return tok.err
	}
	if delim, ok := tok.tok.(json.Delim); ok {
		if rune(delim) == r {
			return nil
		}
	}
	return fmt.Errorf("dymessage: expected %q, but found %q", r, tok.tok)
}

func (tok token) string() (str string, err error) {
	if tok.err != nil {
		err = tok.err
	} else if str, ok := tok.tok.(string); ok {
		return str, nil
	} else {
		err = fmt.Errorf("dymessage: token %q is not a valid string", tok.tok)
	}
	return
}

func (tok token) number() (n json.Number, err error) {
	if tok.err != nil {
		err = tok.err
	} else if n, ok := tok.tok.(json.Number); ok {
		return n, nil
	} else {
		err = fmt.Errorf("dymessage: token %q is not a valid number", tok.tok)
	}
	return
}

func (tok token) boolean() (r bool, err error) {
	if tok.err != nil {
		err = tok.err
	} else if r, ok := tok.tok.(bool); ok {
		return r, nil
	} else {
		err = fmt.Errorf("dymessage: token %q is not a boolean", tok.tok)
	}
	return
}

func (tok token) null() bool {
	return tok.err == nil && tok.tok == nil
}
