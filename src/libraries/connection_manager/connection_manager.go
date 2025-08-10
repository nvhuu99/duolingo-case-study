package connection_manager

import (
	"context"
	"errors"
	"log"
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

func (manager *ConnectionManager) IsNetworkErr(err error) bool {
	return manager.connectionProxy.IsNetworkErr(err)
}

func (manager *ConnectionManager) NotifyNetworkFailure() {
	if manager.resettingTriggered.Load() {
		return
	}
	manager.resettingTriggered.Store(true)

	log.Printf("ConnectionManager(%v): network failure detected, all connection will be discarded for resetting", manager.connectionName())

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

	log.Printf("ConnectionManager(%v): client registered with clientID: %v\n", manager.connectionName(), client.GetClientId())

	id := client.GetClientId()
	manager.clients[id] = client
	manager.clientConnections[id] = manager.makeConnectionAndNotifyIfFails()
	client.connectionManager = manager
}

func (manager *ConnectionManager) RemoveClient(client *Client) {
	id := client.GetClientId()

	log.Printf("ConnectionManager(%v): client is removed, clientID: %v\n", manager.connectionName(), id)

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

	newConn, err := manager.makeConnection()
	if err != nil {
		log.Printf("ConnectionManager(%v): failed to renew client connection - clientId: %v - error: %v\n", manager.connectionName(), client.GetClientId(), err)
		return err
	} else {
		log.Printf("ConnectionManager(%v): client connection renewed - clientID: %v\n", manager.connectionName(), client.GetClientId())
		manager.clientConnections[client.GetClientId()] = newConn
		return nil
	}
}

func (manager *ConnectionManager) discardAllClientConnections() {
	manager.clientsMutex.Lock()
	defer manager.clientsMutex.Unlock()
	defer log.Printf("ConnectionManager(%v): all client connections discarded\n", manager.connectionName())

	for id := range manager.clients {
		if manager.clientConnections[id] != nil {
			go manager.connectionProxy.CloseConnection(manager.clientConnections[id])
		}
		manager.clientConnections[id] = nil
	}
}

func (manager *ConnectionManager) idleUntilNetworkRecoveredUnlessCtxCanceled() error {
	log.Printf("ConnectionManager(%v): waiting for network recorvery\n", manager.connectionName())

	for {
		select {
		case <-manager.ctx.Done():
			return ErrManagerContextCanceled
		default:
			if conn, connErr := manager.makeConnection(); connErr == nil {
				if pingErr := manager.connectionProxy.Ping(conn); pingErr == nil {
					log.Printf("ConnectionManager(%v): network has recovered\n", manager.connectionName())
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
	defer log.Printf("ConnectionManager(%v): all client connections resetted\n", manager.connectionName())

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

func (manager *ConnectionManager) connectionName() string {
	return manager.connectionProxy.ConnectionName()
}
