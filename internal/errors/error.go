package errors

import (
	"errors"
)

// ErrURLAlreadyExists ошибка, возникающая при попытке добавить уже существующий URL.
var ErrURLAlreadyExists = errors.New("URL уже существует")

// ErrURLDeleted ошибка, возникающая при попытке доступа к удаленному URL.
var ErrURLDeleted = errors.New("URL удален")

// ErrURLNotFound ошибка, возникающая при попытке доступа к несуществующему URL.
var ErrURLNotFound = errors.New("URL не найден")
