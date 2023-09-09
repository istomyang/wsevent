package ws

import (
	"context"
	"github.com/gorilla/websocket"
	"wsevent/log"
)

// Client manages WebSocket connections in client end.
type Client interface {
	// Create creates a connection over a http connection and return a Session object.
	Create(addr string, path string) (Session, error)

	Run()
	Close()
}

type client struct {
	ctx      context.Context
	cancel   context.CancelFunc
	config   ClientConfig
	sessions []innerSession
}

func NewClient(ctx context.Context, config ClientConfig) Client {
	ctx, cancel := context.WithCancel(ctx)
	return &client{
		ctx:      ctx,
		cancel:   cancel,
		config:   config,
		sessions: make([]innerSession, 0), // buffer has data loss when panicked.
	}
}

func (c *client) Create(addr string, path string) (Session, error) {
	conn, _, err := websocket.DefaultDialer.DialContext(c.ctx, addr+path, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var se = newSession(c.ctx, conn, path)
	se.(innerSession).Attach()
	c.sessions = append(c.sessions, se.(innerSession))

	log.Debug("ws-client: create session, %v", se)
	return se, nil
}

func (c *client) Run() {
	go func() {
		select {
		case <-c.ctx.Done():
			log.Debug("ws-client: closed for context.Done")
			c.Close()
		}
	}()
}

func (c *client) Close() {
	defer c.cancel()
	for _, se := range c.sessions {
		se.Close()
	}
	log.Debug("ws-client: close")
}

var _ Client = &client{}

type fakeClient struct {
	config FakeClientConfig
}

type FakeClientConfig struct {
	ClientSend <-chan []byte
}

func NewFakeClient(config FakeClientConfig) Client {
	return &fakeClient{config: config}
}

func (f *fakeClient) Create(addr string, path string) (Session, error) {
	log.Debug("ws-fakeClient: create session.")
	return newFakeSession(FakeSessionConfig{ClientSend: f.config.ClientSend}), nil
}

func (f *fakeClient) Run() {
	log.Debug("ws-fakeClient: run.")
}

func (f *fakeClient) Close() {
	log.Debug("ws-fakeClient: close.")
}

var _ Client = &fakeClient{}
