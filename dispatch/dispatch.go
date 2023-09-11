package dispatch

import (
	"context"
	"sync/atomic"
)

// Dispatcher connects Receiver s and Source.
type Dispatcher interface {
	Register(receiver Receiver)
	Connect(source Source)

	Run()
	Close()
}

type dispatcher struct {
	ctx       context.Context
	cancel    context.CancelFunc
	receivers []Receiver
	sources   []Source
	broker    chan Message
	closed    atomic.Bool
}

func NewDispatcher(ctx context.Context) Dispatcher {
	ctx, cancel := context.WithCancel(ctx)
	var dis = dispatcher{
		ctx:       ctx,
		cancel:    cancel,
		receivers: make([]Receiver, 0),
		broker:    make(chan Message),
	}
	return &dis
}

func (d *dispatcher) Register(receiver Receiver) {
	// TODO: reduce slices memory, you can in Run.
	d.receivers = append(d.receivers, receiver)
}

func (d *dispatcher) Connect(source Source) {
	d.sources = append(d.sources, source)
}

func (d *dispatcher) Run() {

	go func() {
		for msg := range d.broker {
			for _, r := range d.receivers {
				r.(innerReceiver).Put(&msg)
			}
		}
	}()

	for _, source := range d.sources {
		source := source
		go func() {
			for msg := range source.(innerSource).Get() {
				if d.closed.Load() {
					break
				}
				d.broker <- msg
			}
		}()
	}

	go func() {
		select {
		case <-d.ctx.Done():
			d.Close()
		}
	}()
}

func (d *dispatcher) Close() {
	defer d.cancel()
	d.closed.Store(true)
}

var _ Dispatcher = &dispatcher{}
