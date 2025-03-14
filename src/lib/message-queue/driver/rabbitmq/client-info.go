package rabbitmq

import (
	"context"
	mq "duolingo/lib/message-queue"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	clientStatusResetting	= "client_resetting"
	clientStatusReady		= "client_ready"
)

// A Client wrapper to help the RabbitMQManager managing a client channel
type clientInfo struct {
	client		mq.Client
	manager		*RabbitMQManager

	id			string
	channel		*amqp.Channel
	channelId	string
	name		string

	status		string
	reset		chan bool

	ctx			context.Context
	cancel		context.CancelFunc
	mu			sync.RWMutex
}

func newClientInfo(name string, manager *RabbitMQManager, client mq.Client) *clientInfo {
	clientCtx, clientCancel := context.WithCancel(manager.ctx)
	
	info := &clientInfo{
		id:	uuid.New().String(),
		name: name,
		client: client,
		manager: manager,
		status: "",
		reset: make(chan bool, 1),
		ctx: clientCtx,
		cancel: clientCancel,
	}

	return info
}

func (c *clientInfo) handleClientChannel() {
	// If this function is return,
	// it means the the connection can not be kept alive.
	// In that case, unregister the client.
	defer c.manager.UnRegisterClient(c.id)
	
	var ch *amqp.Channel
	var err error
	var closedNotifications chan *amqp.Error

	// This func is called below to recreate the client channel.
	recreateChannel := func() bool {
		c.onResetting()
		defer c.onReady()

		firstTry := true
		for {
			select {
			case <-c.ctx.Done():
				return false
			default:
			}
			// Failed to acquire the client channel.
			if ch, err = c.makeChannel(); ch == nil || err != nil {
				// Inform the client of the issue.
				if firstTry {
					go c.client.OnConnectionFailure(err)
					firstTry = false
				}
				continue
			} 
			// Client channel is ready.
			c.mu.Lock()
			c.channel = ch
			c.channelId = uuid.New().String()
			c.mu.Unlock()
			closedNotifications = ch.NotifyClose(make(chan *amqp.Error, 1))
			return true
		}
	}
	// Maintain client channel.
	for {
		select {
		case <-c.ctx.Done():
			return
		case chanErr := <-closedNotifications:
			// Received a channel error (the channel has just closed).
			// Gracefully wait for the connection status to be set by the manager first.  
			time.Sleep(50 * time.Millisecond)
			// If it was not a connection err, and it was actually from the client side,
			// then the client must be informed.
			// It's up to the client to unregister and terminate itself.
			if ! c.manager.IsReConnecting() {
				go c.client.OnClientFatalError(mq.NewError(mq.ClientFatalError, chanErr, "", "", ""))
			}
			// Trigger resetting clients connections.
			go c.triggerReset()
		case <-c.reset:
			if c.isResetting() {
				continue
			}
			// This operation might fail until the connection is ready again.
			// Even when the connection is ready again, it's also possible that
			// this operation fails due to difficulties such as "max channels amount" exceeded,
			// and it need to retry until an available channel is acquired.
			recreateChannel()
		}
	}
}

func (c *clientInfo) makeChannel() (*amqp.Channel, error) {
	var ch *amqp.Channel
	var err error
	retries := 0
	maxRetries := 1000

	for {
		select {
		case <-c.ctx.Done():
			return nil, nil
		default:
		}
		// Wait until the connection is ready
		if ! c.manager.IsReady() {
			time.Sleep(c.manager.opts.GraceTimeOut)
			retries = 0
			continue
		}
		// Check max retries
		if retries == maxRetries {
			break
		}
		retries++
		// Make channel
		c.manager.mu.RLock()
		ch, err = c.manager.conn.Channel()
		c.manager.mu.RUnlock()
		if err == nil {
			return ch, nil
		}
		// Retry gracefully
		if retries > 0 {
			time.Sleep(c.manager.opts.GraceTimeOut)
		}
	}

	return nil, mq.NewError(mq.ConnectionFailure, err, "", "", "")
}

func (c *clientInfo) discardChannel() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.channel != nil && !c.channel.IsClosed() {
		c.channel.Close()
	}
	c.channel = nil
	c.channelId = ""
}

func (c *clientInfo) triggerReset() {
	c.reset <- true
}

func (c *clientInfo) isResetting() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status == clientStatusResetting
}

func (c *clientInfo) onResetting() {
	c.mu.Lock()
	c.status = clientStatusResetting
	c.mu.Unlock()
	c.discardChannel()
}

func (c *clientInfo) onReady() {
	c.mu.Lock()
	c.status = clientStatusReady
	c.mu.Unlock()
}
