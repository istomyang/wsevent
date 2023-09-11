package dispatch

import "context"

type Source interface {
	// Send provides an interface to send to Dispatcher and listen errors flow.
	// You should Decode []byte, get EventKey and encapsulate into Message.
	Send(Message)
}

type innerSource interface {
	Get() <-chan Message
}

type source struct {
	ctx         context.Context
	cancel      context.CancelFunc
	id          int64
	messageChan chan Message
	errorChan   chan error
}

func NewSource() Source {
	return &source{
		messageChan: make(chan Message),
		errorChan:   make(chan error),
	}
}

func (s *source) Get() <-chan Message {
	return s.messageChan
}

func (s *source) Send(message Message) {
	s.messageChan <- message

}

var _ Source = &source{}
var _ innerSource = &source{}
