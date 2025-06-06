package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ClientConnectionManager struct {
	uri                 string
	connectionGraceWait time.Duration

	clients           map[string]*Client
	clientConnections map[string]*mongo.Client

	ctx context.Context
}

func (manager *ClientConnectionManager) SetUri(uri string) {
	manager.uri = uri
	manager.DiscardClients()
}

func (manager *ClientConnectionManager) RegisterClient(client *Client, connectImmediately bool) {
	manager.clients[client.GetClientId()] = client
	client.connectionManager = manager
	if !connectImmediately {
		return
	}
	go func() {
		conn, _ := manager.GetClientConnection(client, true)
		manager.clientConnections[client.GetClientId()] = conn
	}()
}

func (manager *ClientConnectionManager) RemoveClient(client *Client) {
	if conn := manager.clientConnections[client.GetClientId()]; conn != nil {
		conn.Disconnect(manager.ctx)
	}
	delete(manager.clients, client.GetClientId())
	delete(manager.clientConnections, client.GetClientId())
}

func (manager *ClientConnectionManager) DiscardClients() {
	for id := range manager.clients {
		manager.RemoveClient(manager.clients[id])
	}
}

func (manager *ClientConnectionManager) GetClientConnection(
	client *Client,
	forceReconnect bool,
) (
	*mongo.Client,
	error,
) {
	var id = client.GetClientId()
	var currCon = manager.clientConnections[id]
	// connection exists, and a ping is not required
	if currCon != nil && !forceReconnect {
		return currCon, nil
	}
	// if ping fail, try to reset client connection
	if currCon != nil {
		pingCtx, pingCancel := context.WithTimeout(manager.ctx, manager.connectionGraceWait)
		defer pingCancel()
		if pingErr := currCon.Ping(pingCtx, nil); pingErr == nil {
			return currCon, nil
		}
	}
	// reset client connection
	newConn, timeoutErr := manager.makeConnection(client)
	if timeoutErr == nil {
		manager.clientConnections[id] = newConn
	}
	return newConn, timeoutErr
}

func (manager *ClientConnectionManager) makeConnection(client *Client) (*mongo.Client, error) {
	opts := options.Client()
	opts.SetConnectTimeout(client.connectionWait)
	opts.ApplyURI(manager.uri)
	conn, err := mongo.Connect(manager.ctx, opts)
	return conn, err
}
