package rabbitmq

import (
	"context"
	mq "duolingo/lib/message-queue"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type clientInfo struct {
	id			string
	client		mq.Client
	channel		*amqp.Channel
	channelId	string
	done		chan string
	ctx			context.Context
	cancel		context.CancelFunc
	mu			sync.RWMutex
}

type RabbitMQManager struct {
	// non mutex
	conn	*amqp.Connection
	opts	*mq.ManagerOptions

	// mutex
	clients			map[string]*clientInfo
	isReConnecting	bool
	uri				string

	ctx		context.Context
	cancel	context.CancelFunc
	mu		sync.RWMutex
}

func NewRabbitMQManager(ctx context.Context, opts *mq.ManagerOptions) *RabbitMQManager {
	m := RabbitMQManager{}
	m.opts = opts
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.clients = make(map[string]*clientInfo)

	return &m
}

func (m *RabbitMQManager) UseConnection(host string, port string, user string, password string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if user != "" && password != "" {
		m.uri = fmt.Sprintf("amqp://%v:%v@%v:%v/", 
			url.QueryEscape(user), 
			url.QueryEscape(password), 
			host, 
			port,
		)
	} else {
		m.uri = fmt.Sprintf("amqp://%v:%v/", host, port)
	}
}

func (m *RabbitMQManager) Connect() {
	go m.handleReconnect()
}

func (m *RabbitMQManager) RegisterClient(client mq.Client) string {
	clientCtx, clientCancel := context.WithCancel(m.ctx)
	info := &clientInfo{
		id: uuid.New().String(),
		client: client,
		done: make(chan string, 1),
		ctx: clientCtx,
		cancel: clientCancel,
	}

	m.mu.Lock()
	m.clients[info.id] = info
	m.mu.Unlock()

	go m.handleClientChannel(info.id)

	return info.id
}

func (m *RabbitMQManager) UnRegisterClient(id string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, exists := m.clients[id]; exists {
		m.clients[id].done <- id
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

		client.mu.RLock()
		defer client.mu.RUnlock()
		
		if client.channel != nil {
			client.channel.Close()
		}
		client.cancel()
	}
}

func (m *RabbitMQManager) connect() (*amqp.Connection, *mq.Error) {
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
		conn, err = amqp.DialConfig(m.uri, amqp.Config{
			Heartbeat: m.opts.HearBeat,
		})
		if err == nil {
			return conn, nil
		}
		time.Sleep(m.opts.GraceTimeOut)
	}
}

func (m *RabbitMQManager) handleReconnect() {
	var closedNotifications chan *amqp.Error
	// This func is called below to reconnect to the message queue server.
	reConnect := func () bool {
		// Maintain connection status
		m.mu.Lock()
		m.isReConnecting = true
		m.mu.Unlock()
		defer func() {
			m.mu.Lock()
			m.isReConnecting = false
			m.mu.Unlock()
		}()

		var err *mq.Error
		firstTry := true
		for {
			// Failed to connect to the server.
			if m.conn, err = m.connect(); err != nil {
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
			m.mu.RLock()
			for _, info := range m.clients {
				go info.client.OnReConnected()
			}
			m.mu.RUnlock()
			closedNotifications = m.conn.NotifyClose(make(chan *amqp.Error, 1))
			return true
		}
	}
	// First connection.
	if ! reConnect() {
		m.cancel()
		return
	}
	// Maintain the connection.
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-closedNotifications:
			// Received a connection error, and the connection has just closed.
			// First update the connection to nil.
			m.conn = nil
			// Start reconnecting to the server,
			// this operation might fail if a connection cannot be established
			// before the "connection timeout".
			// 
			// Incase of failure, the context is canceled.
			// All clients will also be unregistered. 
			if ! reConnect() {
				m.cancel()
				return
			}
		}
	}
}

func (m *RabbitMQManager) channel(id string) (*amqp.Channel, *mq.Error) {
	var ch *amqp.Channel
	var err error
	retries := 0
	maxRetries := 100

	m.mu.RLock()
	client := m.clients[id]
	m.mu.RUnlock()

	for {
		select {
		case <-m.ctx.Done():
			return nil, nil
		case <-client.ctx.Done():
			return nil, nil
		default:
		}

		if retries > 0 {
			time.Sleep(m.opts.GraceTimeOut)
		}

		m.mu.RLock()
		isReConnecting := m.isReConnecting
		m.mu.RUnlock()
		if isReConnecting {
			retries = 0
			continue
		}

		if retries == maxRetries {
			break
		}
		retries++

		if ch, err = m.conn.Channel(); err == nil {
			return ch, nil
		}
	}

	return nil, mq.NewError(mq.ConnectionFailure, err, "", "", "")
}

func (m *RabbitMQManager) handleClientChannel(id string) {
	// If this function is return,
	// it means the the connection can not be kept alive.
	// In that case, unregister the client.
	defer m.unRegister(id)

	var ch *amqp.Channel
	var err *mq.Error
	var closedNotifications chan *amqp.Error
	m.mu.RLock()
	client := m.clients[id]
	m.mu.RUnlock()
	// This func is called below to recreate the client channel.
	recreateChannel := func() bool {
		firstTry := true
		for {
			select {
			case <-m.ctx.Done():
				return false
			case <-client.ctx.Done():
				return false
			default:
			}
			// Failed to acquire the client channel.
			if ch, err = m.channel(id); err != nil {
				// Inform the client of the issue.
				if firstTry {
					go client.client.OnConnectionFailure(err)
				}
				firstTry = false
				// Stop this operation according to the "keep alive" option.
				if m.opts.KeepAlive {
					time.Sleep(m.opts.HearBeat)
					continue
				}
				return false
			} 
			// Client channel is ready.
			client.mu.Lock()
			defer client.mu.Unlock()
			client.channel = ch
			client.channelId = uuid.New().String()
			closedNotifications = ch.NotifyClose(make(chan *amqp.Error, 1))
			return true
		}
	}
	// First create the client channel.
	if ! recreateChannel() {
		return
	}
	// Maintain client channel.
	for {
		select {
		case <-client.ctx.Done():
			m.unRegister(id)
			return
		case <-client.done:
			// Client requests unregisteration.
			m.unRegister(id)
			return
		case chanErr := <-closedNotifications:
			// Received a channel error, and the channel has just closed.
			// First upate client channel to nil.
			client.mu.Lock()
			client.channel = nil
			client.channelId = ""
			client.mu.Lock()
			// Gracefully wait for the connection status 
			// to be set by the connection handler first.  
			time.Sleep(50 * time.Millisecond) 
			// If it was not a connection err,
			// and it was actually from the client side,
			// the client must be informed.
			// It's up to the client to unregister and terminate itself.
			m.mu.RLock()
			isNotConnectionErr := !m.isReConnecting
			m.mu.RUnlock()
			if isNotConnectionErr {
				go client.client.OnClientFatalError(mq.NewError(mq.ClientFatalError, chanErr, "", "", ""))
			}
			// Other than client side errors, try to recreate the channel.
			// Incase the error was a connection lost, this operation 
			// might fail until the connection is ready again.
			// 
			// Even when the connection is ready again, it's also possible that
			// this operation fails due to difficulties such as "max channels amount" exceeded
			// and it took too long waiting for an available slot.  
			if ! recreateChannel() {
				return
			}
		}
	}
}
