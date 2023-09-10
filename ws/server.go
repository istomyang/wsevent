package ws

import (
	"context"
	"github.com/istomyang/wsevent/log"
	"net/http"
)

// Server manages WebSocket connections in server end.
type Server interface {
	// Create creates a connection over a http connection and return a Session object.
	Create(w http.ResponseWriter, r *http.Request) (Session, error)

	Run()
	Close()
}

type svr struct {
	ctx    context.Context
	cancel context.CancelFunc
	config ServerConfig

	sessions []innerSession
}

func NewServer(ctx context.Context, config ServerConfig) Server {
	ctx, cancel := context.WithCancel(ctx)
	return &svr{
		ctx:      ctx,
		cancel:   cancel,
		config:   config,
		sessions: make([]innerSession, 0), // buffer has data loss when panicked.
	}
}

func (s *svr) Create(w http.ResponseWriter, r *http.Request) (Session, error) {
	conn, err := s.config.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var se = newSession(s.ctx, conn, r.URL.Path)
	se.(innerSession).Attach()
	s.sessions = append(s.sessions, se.(innerSession))

	log.Debug("ws-server: create session, %v.", se)
	return se, nil
}

func (s *svr) Run() {
	go func() {
		select {
		case <-s.ctx.Done():
			log.Debug("ws-server: close by context.Done.")
			s.Close()
		}
	}()
}

func (s *svr) Close() {
	defer s.cancel()
	for _, se := range s.sessions {
		se.Close()
	}
	log.Debug("ws-server: close.")
}

var _ Server = &svr{}

type fakeServer struct {
	config FakeServerConfig
}

type FakeServerConfig struct {
	ClientSend <-chan []byte
}

func NewFakeServer(config FakeServerConfig) Server {
	return &fakeServer{config: config}
}

func (f *fakeServer) Create(w http.ResponseWriter, r *http.Request) (Session, error) {
	log.Debug("ws-fakeServer: create session.")
	return newFakeSession(FakeSessionConfig{ClientSend: f.config.ClientSend}), nil
}

func (f *fakeServer) Run() {
	log.Debug("ws-fakeServer: run.")
}

func (f *fakeServer) Close() {
	log.Debug("ws-fakeServer: close.")
}

var _ Server = &fakeServer{}
