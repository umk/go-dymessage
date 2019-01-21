package json

import (
	"errors"
	"fmt"
	"strings"

	"github.com/umk/go-dymessage/json/internal/impl"
)

// -----------------------------------------------------------------------------
// Token query methods

func (dc *decoder) accept(tk impl.TokenKind) error {
	if dc.lx.Err != nil {
		return dc.lx.Err
	} else if dc.lx.Tok.Kind != tk {
		return errors.New(dc.createErrorMessage(tk))
	}
	dc.lx.Next()
	return nil
}

func (dc *decoder) acceptSeq(tk ...impl.TokenKind) (err error) {
	for _, t := range tk {
		if err = dc.accept(t); err != nil {
			return
		}
	}
	return
}

func (dc *decoder) tryAccept(tk impl.TokenKind) (accepted bool) {
	accepted = dc.lx.Err == nil && dc.lx.Tok.Kind == tk
	if accepted {
		dc.lx.Next()
	}
	return
}

func (dc *decoder) tryAcceptAny(tk ...impl.TokenKind) (accepted bool) {
	if dc.lx.Err != nil {
		return false
	}
	for _, t := range tk {
		if t == dc.lx.Tok.Kind {
			dc.lx.Next()
			return true
		}
	}
	return false
}

func (dc *decoder) acceptValue(tk impl.TokenKind) (str string, err error) {
	if err = dc.accept(tk); err != nil {
		return
	}
	return dc.lx.Tok.Value, nil
}

func (dc *decoder) acceptBool() (b bool, err error) {
	if err = dc.lx.Err; err != nil {
		return
	} else if tk := dc.lx.Tok.Kind; tk != impl.TkTrue && tk != impl.TkFalse {
		err = errors.New(dc.createErrorMessage(impl.TkTrue, impl.TkFalse))
	} else {
		b = (tk == impl.TkTrue)
		dc.lx.Next()
	}
	return
}

func (dc *decoder) probably(tk impl.TokenKind) bool {
	return dc.lx.Err == nil && dc.lx.Tok.Kind == tk
}

// -----------------------------------------------------------------------------
// Helper methods

// getCurrentToken gets the current token, just like it was represented in the
// original JSON input.
func (dc *decoder) getCurrentToken() string {
	switch dc.lx.Tok.Kind {
	case impl.TkString, impl.TkNumber, impl.TkTrue, impl.TkFalse, impl.TkNull:
		return dc.lx.Tok.Value
	default:
		return dc.lx.Tok.Kind.String()
	}
}

func (dc *decoder) createErrorMessage(tk ...impl.TokenKind) string {
	if len(tk) == 0 {
		return fmt.Sprintf("dymessage: %v: unexpected token %q",
			dc.lx.Tok.Pos, dc.getCurrentToken())
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("dymessage: %v: expected ", dc.lx.Tok.Pos))
	n := len(tk)
	sb.WriteString(tk[0].MessageString())
	if n > 1 {
		for i := 1; i < n-1; i++ {
			sb.WriteString(", ")
			sb.WriteString(tk[i].MessageString())
		}
		sb.WriteString(" or ")
		sb.WriteString(tk[n-1].MessageString())
	}
	sb.WriteString(", but found ")
	sb.WriteString(fmt.Sprintf("%q", dc.getCurrentToken()))
	return sb.String()
}
