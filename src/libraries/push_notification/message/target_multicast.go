package message

import (
	"errors"
	"slices"
)

var (
	ErrMulticastTargetInadequate = errors.New("multicast device tokens, and platforms must be specified")
)

type MulticastTarget struct {
	Platforms    []Platform
	DeviceTokens []string
}

func (t *MulticastTarget) Validate() error {
	if len(t.DeviceTokens) == 0 || slices.Contains(t.DeviceTokens, "") ||
		len(t.Platforms) == 0 {
		return ErrMulticastTargetInadequate
	}
	return nil
}
