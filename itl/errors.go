package itl

import (
	"errors"
)

var ErrTooBig = errors.New("object bigger than encoded size")
var ErrInvalidHeader = errors.New("invalid header")
var ErrUnexpectedObject = errors.New("unexpected object")
var ErrUnknownVersion = errors.New("unknown file format version")
