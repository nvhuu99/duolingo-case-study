package message

import (
	"errors"
	"time"
)

var (
	ErrMessageRequireParamsMissing = errors.New("title, body must not be empty string")
)

type Message struct {
	Title       string
	Body        string
	Icon        string
	Sound       string
	Expiration  time.Duration
	CollapseKey string
	Priority    Priority
	Visibility  Visibility
}

func (n *Message) Validate() error {
	if n.Title == "" || n.Body == "" {
		return ErrMessageRequireParamsMissing
	}
	return nil
}
