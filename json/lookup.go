package json

import (
	"errors"
	"fmt"
	"strings"
)

// -----------------------------------------------------------------------------
// Token query methods

func (dc *decoder) accept(tk tokenKind) error {
	if dc.lx.err != nil {
		return dc.lx.err
	} else if dc.lx.tok.kind != tk {
		return errors.New(dc.createErrorMessage(tk))
	}
	dc.lx.next()
	return nil
}

func (dc *decoder) acceptSeq(tk ...tokenKind) (err error) {
	for _, t := range tk {
		if err = dc.accept(t); err != nil {
			return
		}
	}
	return
}

func (dc *decoder) tryAccept(tk tokenKind) (accepted bool) {
	accepted = dc.lx.err == nil && dc.lx.tok.kind == tk
	if accepted {
		dc.lx.next()
	}
	return
}

func (dc *decoder) tryAcceptAny(tk ...tokenKind) (accepted bool) {
	if dc.lx.err != nil {
		return false
	}
	for _, t := range tk {
		if t == dc.lx.tok.kind {
			dc.lx.next()
			return true
		}
	}
	return false
}

func (dc *decoder) acceptValue(tk tokenKind) (str string, err error) {
	if err = dc.accept(tk); err != nil {
		return
	}
	return dc.lx.tok.value, nil
}

func (dc *decoder) acceptBool() (b bool, err error) {
	if err = dc.lx.err; err != nil {
		return
	} else if tk := dc.lx.tok.kind; tk != tkTrue && tk != tkFalse {
		err = errors.New(dc.createErrorMessage(tkTrue, tkFalse))
	} else {
		b = (tk == tkTrue)
		dc.lx.next()
	}
	return
}

func (dc *decoder) probably(tk tokenKind) bool {
	return dc.lx.err == nil && dc.lx.tok.kind == tk
}

// -----------------------------------------------------------------------------
// Helper methods

// getCurrentToken gets the current token, just like it was represented in the
// original JSON input.
func (dc *decoder) getCurrentToken() string {
	switch dc.lx.tok.kind {
	case tkString, tkNumber, tkTrue, tkFalse, tkNull:
		return dc.lx.tok.value
	default:
		return dc.lx.tok.kind.String()
	}
}

func (dc *decoder) createErrorMessage(tk ...tokenKind) string {
	if len(tk) == 0 {
		return fmt.Sprintf("dymessage: %v: unexpected token %q",
			dc.lx.tok.pos, dc.getCurrentToken())
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("dymessage: %v: expected ", dc.lx.tok.pos))
	n := len(tk)
	sb.WriteString(tk[0].messageString())
	if n > 1 {
		for i := 1; i < n-1; i++ {
			sb.WriteString(", ")
			sb.WriteString(tk[i].messageString())
		}
		sb.WriteString(" or ")
		sb.WriteString(tk[n-1].messageString())
	}
	sb.WriteString(", but found ")
	sb.WriteString(fmt.Sprintf("%q", dc.getCurrentToken()))
	return sb.String()
}
