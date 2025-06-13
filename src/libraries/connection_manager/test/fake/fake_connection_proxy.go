package fake

import (
	"errors"
	"sync/atomic"
)

var (
	ErrFakeNetworkFailure = errors.New("fake connection network failure")
)

type FakeConnectionProxy struct {
	networkUp atomic.Bool
}

func NewFakeConnectionProxy() *FakeConnectionProxy {
	f := &FakeConnectionProxy{}
	f.networkUp.Store(true)
	return f
}

func (f *FakeConnectionProxy) SetArgsPanicIfInvalid(args any) {
}

func (f *FakeConnectionProxy) GetConnection() (any, error) {
	if !f.networkUp.Load() {
		return nil, errors.New("")
	}
	return new(FakeConnection), nil
}

func (f *FakeConnectionProxy) Ping(connection any) error {
	if !f.networkUp.Load() {
		return ErrFakeNetworkFailure
	}
	if _, ok := connection.(*FakeConnection); !ok {
		return ErrFakeNetworkFailure
	}
	return nil
}

func (f *FakeConnectionProxy) CloseConnection(connection any) {
}

func (f *FakeConnectionProxy) IsNetworkErr(err error) bool {
	return err == ErrFakeNetworkFailure
}

func (f *FakeConnectionProxy) IsNetworkUp() bool {
	return f.networkUp.Load()
}

func (f *FakeConnectionProxy) SimulateNetworkFailure() {
	f.networkUp.Store(false)
}

func (f *FakeConnectionProxy) SimulateNetworkRecovery() {
	f.networkUp.Store(true)
}
