package mailbus

import (
	"bytes"
	"errors"
	"fmt"
)

const (
	ErrInvalid      = "invalid"
	ErrUnauthorized = "unauthorized"
	ErrForbidden    = "forbidden"
	ErrNotFound     = "not_found"
	ErrConflict     = "conflict"
	ErrInternal     = "internal"
)

type Error struct {
	Code    string
	Message string
	Op      string
	Err     error
}

func ErrorCode(err error) string {
	var e *Error
	if err == nil {
		return ""
	} else if errors.As(err, &e) && e.Code != "" {
		return e.Code
	} else if e.Err != nil {
		return ErrorCode(e.Err)
	}

	return ErrInternal
}

func ErrorMessage(err error) string {
	var e *Error
	if err == nil {
		return ""
	} else if errors.As(err, &e) && e.Message != "" {
		return e.Message
	} else if e.Err != nil {
		return ErrorMessage(e.Err)
	}

	return "An internal error has occurred."
}

func (e *Error) Error() string {
	var buf bytes.Buffer

	if e.Op != "" {
		fmt.Fprintf(&buf, "%s: ", e.Op)
	}

	if e.Err != nil {
		buf.WriteString(e.Err.Error())
	} else {
		if e.Code != "" {
			fmt.Fprintf(&buf, "<%s> ", e.Code)
		}
		buf.WriteString(e.Message)
	}

	return buf.String()
}
