package merge

import (
	"context"
	"sync/atomic"
	"time"
)

type Merge interface {
	// UpdateOnline provides change time.Duration on running time.
	UpdateOnline(duration time.Duration)
	// Allowed tells your goroutine to go on or return directly.
	Allowed(key Key) bool
	// Clear clears local storage of task keys.
	Clear()
	Run()
	Close()
}

type void struct{}

var noop void

type Key string

type merge struct {
	ctx     context.Context
	cancel  context.CancelFunc
	storage map[Key]void
	ticker  *time.Ticker
	v       uint32
}

func NewMerge(ctx context.Context, interval time.Duration) Merge {
	ctx, cancel := context.WithCancel(ctx)
	return &merge{
		ctx:     ctx,
		cancel:  cancel,
		storage: make(map[Key]void),
		ticker:  time.NewTicker(interval),
	}
}

func (m *merge) UpdateOnline(duration time.Duration) {
	if duration < time.Millisecond {
		return
	}
	m.ticker.Reset(duration)
}

func (m *merge) Allowed(key Key) bool {

	v := &m.v

lo1:
	for {
		if atomic.LoadUint32(v) == 0 {
			atomic.StoreUint32(v, 1)
			_, has := m.storage[key]
			atomic.StoreUint32(v, 0)
			if has {
				return false
			}
			break lo1
		}
	}

lo2:
	for {
		if atomic.LoadUint32(v) == 0 {
			atomic.StoreUint32(v, 1)
			m.storage[key] = noop
			atomic.StoreUint32(v, 0)
			break lo2
		}
	}

	return true
}

func (m *merge) Clear() {
	var newStorage = make(map[Key]void)

	v := &m.v

lo:
	for {
		if atomic.LoadUint32(v) == 0 {
			atomic.StoreUint32(v, 1)
			m.storage = newStorage
			atomic.StoreUint32(v, 0)
			break lo
		}
	}
}

func (m *merge) Run() {
	go func() {
		for range m.ticker.C {
			m.Clear()
		}
	}()

	go func() {
		select {
		case <-m.ctx.Done():
			m.Close()
		}
	}()
}

func (m *merge) Close() {
	defer m.cancel()
	m.ticker.Stop()
}

var _ Merge = &merge{}
