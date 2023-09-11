package dispatch

import "sync"

type Receiver interface {
	// Get gets Message from Source with EventKey to filter.
	Get() <-chan Message
	Update(eventKeys []EventKey)
}

type innerReceiver interface {
	// Put puts Message into Receiver.
	// Receiver should do Filter.
	// message use nil can reduce unnecessary copy.
	Put(message *Message)
}

type receiver struct {
	messageChan  chan Message
	eventKeysMap map[EventKey]void
	mut          sync.RWMutex
}

func NewReceiver() Receiver {
	return &receiver{
		eventKeysMap: make(map[EventKey]void),
		messageChan:  make(chan Message),
	}
}

func (r *receiver) Get() <-chan Message {
	return r.messageChan
}

func (r *receiver) Update(eventKeys []EventKey) {
	var newMap = make(map[EventKey]void)
	for _, key := range eventKeys {
		newMap[key] = noop
	}
	r.mut.Lock()
	r.eventKeysMap = newMap
	r.mut.Unlock()
}

func (r *receiver) Put(message *Message) {
	// fast skip.
	r.mut.RLocker()
	if _, has := r.eventKeysMap[message.Key]; !has {
		return
	}
	r.mut.RUnlock()

	r.messageChan <- *message
}

var _ Receiver = &receiver{}
var _ innerReceiver = &receiver{}
