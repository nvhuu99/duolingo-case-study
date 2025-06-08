package fake

import (
	"context"
	"duolingo/libraries/connection_manager"
	"errors"
	"log"
	"sync/atomic"
	"time"
)

var (
	ErrFakeNetworkFailure = errors.New("fake connection network failure")
)

type FakeConnectionProxy struct {
	networkUp        atomic.Bool
}

func NewFakeConnectionProxy() *FakeConnectionProxy {
	f := &FakeConnectionProxy{}
	f.networkUp.Store(true)
	return f
}

func (f *FakeConnectionProxy) CreateConnection(args *connection_manager.ConnectArgs) (any, error) {
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

func (f *FakeConnectionProxy) IsNetworkError(err error) bool {
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

func (f *FakeConnectionProxy) SimulateNetworkFailureWithInterval(
	ctx context.Context, 
	interval time.Duration,
) {
	failureTicker := time.Tick(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-failureTicker:
				// Toggle network state
				log.Println("im here though")
				if f.networkUp.Load() {
					f.SimulateNetworkFailure()
				} else {
					f.SimulateNetworkRecovery()
				}
			}
		}
	}()
}