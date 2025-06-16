package mongodb

import "errors"

var (
	ErrInvalidCommandType = errors.New("received invalid command type. expect command of mongodb driver")
)
