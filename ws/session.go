package ws

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"sync/atomic"
	"wsevent/log"
)

// Session 's ownership belongs to ws.Client or ws.Server.
type Session interface {
	// Receive gets message from ws client.
	Receive() <-chan []byte
	// Send sends message to ws client.
	Send(data []byte) error
}

type innerSession interface {
	Close()
	// Attach attaches ws into a http connection.
	Attach()
}

// Session holds a chat session context between client and server.
// You can use Receive and Send functions conveniently.
type session struct {
	conn        *websocket.Conn
	path        string
	receiveChan chan []byte
	sendChan    chan []byte
	closed      atomic.Bool
	ctx         context.Context
	cancel      context.CancelFunc
}

// newSession create a Session.
// It's recommend to create Session use Client.Create.
func newSession(ctx context.Context, conn *websocket.Conn, path string) Session {
	ctx, cancel := context.WithCancel(ctx)
	return &session{
		ctx:         ctx,
		cancel:      cancel,
		conn:        conn,
		path:        path,
		receiveChan: make(chan []byte),
		sendChan:    make(chan []byte),
	}
}

var _ Session = &session{}
var _ innerSession = &session{}

func (s *session) Attach() {
	go func() {
		select {
		case <-s.ctx.Done():
			log.Debug("ws-session: closed by context.Done.")
			s.Close()
		}
	}()

	go func() {
		for {
			messageType, data, err := s.conn.ReadMessage()
			if err != nil {
				log.Error(err)
				break
			}
			log.Debug("ws-session: conn.ReadMessage, %v", string(data))
			if messageType == websocket.CloseMessage {
				break
			}
			s.receiveChan <- data
		}
	}()

	go func() {
		for send := range s.sendChan {
			if err := s.conn.WriteMessage(websocket.BinaryMessage, send); err != nil {
				log.Error(err)
				break
			}
			log.Debug("ws-session: conn.WriteMessage, %v", string(send))
		}
	}()
}

func (s *session) Close() {
	if s.closed.Load() {
		return
	}
	s.closed.Store(true)

	defer s.cancel()

	close(s.sendChan)
	close(s.receiveChan)

	log.Debug("ws-session: close")
}

func (s *session) Receive() <-chan []byte {
	// Read closed can't cause panic
	return s.receiveChan
}

func (s *session) Send(data []byte) error {
	// Put data into closed channel cause panic.
	// The reason for using atomic instead of scalar type is not ensure happens-before in concurrency.
	if s.closed.Load() {
		return errors.New("session is closed")
	}
	s.sendChan <- data

	log.Debug("ws-session: Send, %v", string(data))
	return nil
}

type fakeSession struct {
	config FakeSessionConfig
}

type FakeSessionConfig struct {
	ClientSend <-chan []byte
}

func newFakeSession(config FakeSessionConfig) Session {
	return &fakeSession{
		config: config,
	}
}

func (f *fakeSession) Receive() <-chan []byte {
	return f.config.ClientSend
}

func (f *fakeSession) Send(data []byte) error {
	log.Debug("session-send: %s", string(data))
	return nil // discard
}

var _ Session = &fakeSession{}
