package subscribe

import (
	"context"
	"github.com/istomyang/wsevent/log"
	"github.com/istomyang/wsevent/ws"
)

type wsSubscriber struct {
	ctx    context.Context
	cancel context.CancelFunc
	config WsConfig

	session ws.Session
	err     error
}

type WsConfig struct {
	Addr string
	Path string
}

func NewWsSubscriber(ctx context.Context, config WsConfig) Subscribe {
	ctx, cancel := context.WithCancel(ctx)
	client := ws.NewClient(ctx, ws.ClientConfig{})
	session, err := client.Create(config.Addr, config.Path)
	return &wsSubscriber{
		ctx:     ctx,
		cancel:  cancel,
		config:  config,
		session: session,
		err:     err,
	}
}

func (w *wsSubscriber) Get() (<-chan []byte, error) {
	return w.session.Receive(), nil
}

func (w *wsSubscriber) Run() error {
	go func() {
		select {
		case <-w.ctx.Done():
			log.Debug("wsSubscriber: closed by context.Done")
			if err := w.Close(); err != nil {
				log.Error(err)
			}
		}
	}()

	return w.err
}

func (w *wsSubscriber) Close() error {
	defer w.cancel()
	w.session = nil

	log.Debug("wsSubscriber: closed")
	return nil
}

var _ Subscribe = &wsSubscriber{}
