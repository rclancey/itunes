package mdb

import (
	"errors"
)

var ErrTooBig = errors.New("object bigger than encoded size")
var ErrInvalidHeader = errors.New("invalid library header")
var ErrUnexpectedObject = errors.New("unexpected object")
