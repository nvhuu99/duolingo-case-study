package connection_manager

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrManagerContextCanceled = errors.New("connection manager operation canceled by context")
)

type ConnectionManager struct {
	connectionProxy ConnectionProxy

	connectionRetryWait time.Duration
	clients             map[string]*Client
	clientConnections   map[string]any

	ctx                context.Context
	clientsMutex       sync.Mutex
	resettingTriggered atomic.Bool
}

func (manager *ConnectionManager) IsNetworkError(err error) bool {
	return manager.connectionProxy.IsNetworkErr(err)
}

func (manager *ConnectionManager) NotifyNetworkFailure() {
	if manager.resettingTriggered.Load() {
		return
	}
	go func() {
		manager.resettingTriggered.Store(true)
		defer manager.resettingTriggered.Store(false)

		manager.discardAllClientConnections() // avoid further accesses
		if err := manager.idleUntilNetworkRecoveredUnlessCtxCanceled(); err != nil {
			return
		}
		manager.resetAllConnections()
	}()
}

func (manager *ConnectionManager) RegisterClient(client *Client) {
	manager.clientsMutex.Lock()
	defer manager.clientsMutex.Unlock()

	id := client.GetClientId()
	manager.clients[id] = client
	manager.clientConnections[id] = manager.makeConnectionAndNotifyIfFails()
	client.connectionManager = manager
}

func (manager *ConnectionManager) RemoveClient(client *Client) {
	id := client.GetClientId()

	manager.clientsMutex.Lock()
	defer manager.clientsMutex.Unlock()

	conn := manager.clientConnections[id]
	go manager.connectionProxy.CloseConnection(conn)

	delete(manager.clients, id)
	delete(manager.clientConnections, id)
}

func (manager *ConnectionManager) RemoveAllClients() {
	for id := range manager.clients {
		manager.RemoveClient(manager.clients[id])
	}
}

func (manager *ConnectionManager) GetClientConnection(client *Client) any {
	manager.clientsMutex.Lock()
	defer manager.clientsMutex.Unlock()
	return manager.clientConnections[client.GetClientId()]
}

func (manager *ConnectionManager) discardAllClientConnections() {
	manager.clientsMutex.Lock()
	defer manager.clientsMutex.Unlock()

	for id := range manager.clients {
		if manager.clientConnections[id] != nil {
			go manager.connectionProxy.CloseConnection(manager.clientConnections[id])
		}
		manager.clientConnections[id] = nil
	}
}

func (manager *ConnectionManager) idleUntilNetworkRecoveredUnlessCtxCanceled() error {
	for {
		select {
		case <-manager.ctx.Done():
			return ErrManagerContextCanceled
		default:
			if conn, connErr := manager.makeConnection(); connErr == nil {
				if pingErr := manager.connectionProxy.Ping(conn); pingErr == nil {
					return nil
				}
			}
			time.Sleep(manager.connectionRetryWait)
		}
	}
}

func (manager *ConnectionManager) resetAllConnections() {
	manager.clientsMutex.Lock()
	defer manager.clientsMutex.Unlock()

	for id := range manager.clients {
		conn, _ := manager.makeConnection()
		manager.clientConnections[id] = conn
	}
}

func (manager *ConnectionManager) makeConnectionAndNotifyIfFails() any {
	conn, err := manager.makeConnection()
	if err != nil {
		manager.NotifyNetworkFailure()
		return nil
	}
	return conn
}

func (manager *ConnectionManager) makeConnection() (any, error) {
	return manager.connectionProxy.GetConnection()
}
