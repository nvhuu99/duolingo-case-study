package message

import (
	"errors"
	"slices"
)

var (
	ErrMulticastDeviceTokenMissing = errors.New("multicast device tokens must not be empty string")
)

type MulticastTarget struct {
	Platform
	DeviceTokens []string
}

func (t *MulticastTarget) Validate() error {
	if len(t.DeviceTokens) == 0 || slices.Contains(t.DeviceTokens, "") {
		return ErrMulticastDeviceTokenMissing
	}
	return nil
}
