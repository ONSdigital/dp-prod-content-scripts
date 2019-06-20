package errs

import (
	"github.com/ONSdigital/log.go/log"
)

type Error struct {
	Data        log.Data
	Message     string
	OriginalErr error
}

// Construct a new error
func New(message string, err error, data log.Data) Error {
	return Error{
		Message:     message,
		Data:        data,
		OriginalErr: err,
	}
}

func (e Error) Error() string {
	return e.Message
}
