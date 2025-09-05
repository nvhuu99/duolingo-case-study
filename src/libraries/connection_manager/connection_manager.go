package connection_manager

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	events "duolingo/libraries/events/facade"
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

func (manager *ConnectionManager) IsNetworkErr(err error) bool {
	return manager.connectionProxy.IsNetworkErr(err)
}

func (manager *ConnectionManager) NotifyNetworkFailure() {
	if manager.resettingTriggered.Load() {
		return
	}
	manager.resettingTriggered.Store(true)

	events.Emit(manager.ctx, "connection_manager", nil, map[string]any{
		"message":         "network failure detected, resetting connections",
		"connection_name": manager.connectionProxy.ConnectionName(),
	})

	go func() {
		defer manager.resettingTriggered.Store(false)

		manager.discardAllClientConnections() // avoid further accesses
		if err := manager.idleUntilNetworkRecoveredUnlessCtxCanceled(); err != nil {
			return
		}
		manager.resetAllConnectionsForNetworkRecovery()
	}()
}

func (manager *ConnectionManager) RegisterClient(client *Client) {
	manager.clientsMutex.Lock()
	defer manager.clientsMutex.Unlock()

	events.Emit(manager.ctx, "connection_manager", nil, map[string]any{
		"message":         fmt.Sprintf("client registered - ID: %v", client.GetClientId()),
		"connection_name": manager.connectionProxy.ConnectionName(),
	})

	id := client.GetClientId()
	manager.clients[id] = client
	manager.clientConnections[id] = manager.makeConnectionAndNotifyIfFails()
	client.connectionManager = manager
}

func (manager *ConnectionManager) RemoveClient(client *Client) {
	id := client.GetClientId()

	events.Emit(manager.ctx, "connection_manager", nil, map[string]any{
		"message":         fmt.Sprintf("client is removed - ID: %v", client.GetClientId()),
		"connection_name": manager.connectionProxy.ConnectionName(),
	})

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

func (manager *ConnectionManager) RenewClientConnection(client *Client) error {
	manager.clientsMutex.Lock()
	defer manager.clientsMutex.Unlock()

	currConn := manager.clientConnections[client.GetClientId()]
	if currConn != nil {
		manager.connectionProxy.CloseConnection(currConn)
	}
	manager.clientConnections[client.GetClientId()] = nil

	evt := events.Start(manager.ctx, "connection_manager", map[string]any{
		"connection_name": manager.connectionProxy.ConnectionName(),
	})

	newConn, err := manager.makeConnection()
	if err != nil {
		events.Failed(evt, err, map[string]any{
			"message": fmt.Sprintf("failed to renew client connection - ID: %v", client.GetClientId()),
		})
		return err
	} else {
		events.Succeeded(evt, map[string]any{
			"message": fmt.Sprintf("client connection renewed - ID: %v", client.GetClientId()),
		})
		manager.clientConnections[client.GetClientId()] = newConn
		return nil
	}
}

func (manager *ConnectionManager) discardAllClientConnections() {
	manager.clientsMutex.Lock()
	defer manager.clientsMutex.Unlock()

	defer events.Emit(manager.ctx, "connection_manager", nil, map[string]any{
		"message":         "all client connections discarded",
		"connection_name": manager.connectionProxy.ConnectionName(),
	})

	for id := range manager.clients {
		if manager.clientConnections[id] != nil {
			go manager.connectionProxy.CloseConnection(manager.clientConnections[id])
		}
		manager.clientConnections[id] = nil
	}
}

func (manager *ConnectionManager) idleUntilNetworkRecoveredUnlessCtxCanceled() error {
	events.Emit(manager.ctx, "connection_manager", nil, map[string]any{
		"message":         "waiting for network recorvery",
		"connection_name": manager.connectionProxy.ConnectionName(),
	})
	defer events.Emit(manager.ctx, "connection_manager", nil, map[string]any{
		"message":         "network has recovered",
		"connection_name": manager.connectionProxy.ConnectionName(),
	})

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

func (manager *ConnectionManager) resetAllConnectionsForNetworkRecovery() {
	manager.clientsMutex.Lock()
	defer manager.clientsMutex.Unlock()

	defer events.Emit(manager.ctx, "connection_manager", nil, map[string]any{
		"message":         "all client connections resetted",
		"connection_name": manager.connectionProxy.ConnectionName(),
	})

	for id := range manager.clients {
		// Error is ignored, since this function is called only when the network has just recovered.
		// If the connection failed here, the manager will be notified later.
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
	return manager.connectionProxy.MakeConnection()
}
