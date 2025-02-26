package errors

import (
	"errors"
)

var ErrURLAlreadyExists = errors.New("URL already exists")
var ErrURLDeleted = errors.New("URL deleted")
