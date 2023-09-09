package merge

import (
	"context"
	"sync"
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
	mut     sync.RWMutex
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
	m.mut.RLock()
	_, has := m.storage[key]
	m.mut.RUnlock()
	if has {
		return false
	}

	m.mut.Lock()
	m.storage[key] = noop
	m.mut.Unlock()

	return true
}

func (m *merge) Clear() {
	var newStorage = make(map[Key]void)
	m.mut.Lock()
	m.storage = newStorage
	m.mut.Unlock()
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
