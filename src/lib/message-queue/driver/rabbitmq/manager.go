package rabbitmq

import (
	"context"
	mq "duolingo/lib/message-queue"
	"fmt"
	"net/url"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	statusReconnecting = "status_reconnecting"
	statusReady        = "status_ready"
)

type RabbitMQManager struct {
	uri     string
	opts    *mq.ManagerOptions
	conn    *amqp.Connection
	clients map[string]*clientInfo

	status string
	reset  chan bool

	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

func NewRabbitMQManager(ctx context.Context) *RabbitMQManager {
	m := RabbitMQManager{}
	m.opts = mq.DefaultManagerOptions()
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.clients = make(map[string]*clientInfo)
	m.reset = make(chan bool, 1)
	m.status = ""

	return &m
}

func (m *RabbitMQManager) WithOptions(opts *mq.ManagerOptions) *mq.ManagerOptions {
	if opts == nil {
		opts = mq.DefaultManagerOptions()
	}
	if m.opts != nil {
		m.opts = opts
	}
	return m.opts
}

func (m *RabbitMQManager) UseConnection(host, port, user, password string) {
	m.mu.Lock()
	last := m.uri
	uri := ""
	if user != "" && password != "" {
		uri = fmt.Sprintf("amqp://%v:%v@%v:%v/", url.QueryEscape(user), url.QueryEscape(password), host, port)
	} else {
		uri = fmt.Sprintf("amqp://%v:%v/", host, port)
	}
	m.uri = uri
	m.mu.Unlock()

	if last != uri && !m.IsReConnecting() {
		m.reset <- true
	}
}

func (m *RabbitMQManager) Connect() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.opts == nil || m.uri == "" {
		return mq.NewError(mq.ManagerConfigMissing, nil, "", "", "")
	}

	go m.handleReconnect()

	return nil
}

func (m *RabbitMQManager) Disconnect() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.conn != nil {
		m.conn.Close()
	}
	m.cancel()
}

func (m *RabbitMQManager) IsReady() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.status == statusReady && m.conn != nil
}

func (m *RabbitMQManager) IsReConnecting() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.status == statusReconnecting && m.conn == nil
}

func (m *RabbitMQManager) RegisterClient(name string, client mq.Client) string {
	info := newClientInfo(name, m, client)

	m.mu.Lock()
	m.clients[info.id] = info
	m.mu.Unlock()

	go info.handleClientChannel()

	return info.id
}

func (m *RabbitMQManager) UnRegisterClient(id string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, exists := m.clients[id]; exists {
		m.unRegister(id)
	}
}

func (m *RabbitMQManager) GetClientConnection(id string) (any, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.clients[id]; !exists {
		return nil, ""
	}

	m.clients[id].mu.RLock()
	defer m.clients[id].mu.RUnlock()

	return m.clients[id].channel, m.clients[id].channelId
}

func (m *RabbitMQManager) unRegister(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, found := m.clients[id]; found {
		delete(m.clients, id)
		client.discardChannel()
		client.cancel()
	}
}

func (m *RabbitMQManager) connect() (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error
	connectDeadline := time.After(m.opts.ConnectionTimeOut)
	for {
		select {
		case <-m.ctx.Done():
			return nil, nil
		case <-connectDeadline:
			return nil, mq.NewError(mq.ConnectionTimeOut, err, "", "", "")
		default:
		}

		m.mu.RLock()
		uri := m.uri
		m.mu.RUnlock()

		conn, err = amqp.DialConfig(uri, amqp.Config{
			Heartbeat: m.opts.HearBeat,
			Dial:      amqp.DefaultDial(m.opts.ConnectionTimeOut),
		})

		if err == nil {
			return conn, nil
		}

		time.Sleep(m.opts.GraceTimeOut)
	}
}

func (m *RabbitMQManager) onReConnecting() {
	m.mu.Lock()
	if m.conn != nil && !m.conn.IsClosed() {
		m.conn.Close()
	}
	m.conn = nil
	m.status = statusReconnecting
	m.mu.Unlock()
	// Signal the clients to drop their channels
	for _, client := range m.clients {
		client.discardChannel()
	}
}

func (m *RabbitMQManager) onReConnected() {
	m.mu.Lock()
	m.status = statusReady
	m.mu.Unlock()
	// Signal the clients to acquire new channels
	for _, client := range m.clients {
		go client.triggerReset()
	}
}

func (m *RabbitMQManager) handleReconnect() {
	defer m.Disconnect()

	var closedNotifications chan *amqp.Error
	var conn *amqp.Connection

	// This func is called below to reconnect to the message queue server.
	reConnect := func() bool {
		m.onReConnecting()
		defer m.onReConnected()

		var err error
		firstTry := true
		for {
			select {
			case <-m.ctx.Done():
				return false
			default:
			}
			// Failed to connect to the server.
			if conn, err = m.connect(); conn == nil || err != nil {
				// Inform the client of the issue.
				if firstTry {
					m.mu.RLock()
					for _, info := range m.clients {
						go info.client.OnConnectionFailure(err)
					}
					m.mu.RUnlock()
				}
				firstTry = false
				// Stop this operation according to the "keep alive" option.
				if m.opts.KeepAlive {
					time.Sleep(m.opts.HearBeat)
					continue
				}
				return false
			}
			// Connection is ready, inform the clients,
			// and start listening to the next connection closed.
			m.mu.Lock()
			m.conn = conn
			for _, info := range m.clients {
				go info.client.OnReConnected()
			}
			closedNotifications = m.conn.NotifyClose(make(chan *amqp.Error, 1))
			m.mu.Unlock()
			return true
		}
	}
	// Maintain the connection.
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-closedNotifications:
			// Received a connection error, and the connection has just closed.
			// Send to reset channel to trigger resetting the connection.
			go func() {
				m.reset <- true
			}()
		case <-m.reset:
			if m.IsReConnecting() {
				continue
			}
			// Start reconnecting to the server,
			// this operation might fail if a connection cannot be established
			// before the "connection timeout".
			//
			// Incase of failure, the context is canceled.
			// All clients will also be unregistered.
			if !reConnect() {
				m.cancel()
			}
		}
	}
}
