package wsevent

import (
	"fmt"
	"github.com/istomyang/wsevent/log"
	"github.com/istomyang/wsevent/ws"
)

type Register interface {
}

type innerRegister interface {
	// Dispatch dispatches an event from upstream subscribe.Subscribe to ws.Session.
	//
	// Notice that Dispatch should filter inactive events in a fast path.
	// message is a pointer to reduce unnecessary copy.
	Dispatch(key EventKey, message *[]byte) error

	Run() error
	Close() error
}

type RegisterConfig struct {
	// HandlerMap defines handler for specific event key.
	// EventKey comes from subscribe.Subscribe to Dispatcher, represent data change notification.
	//
	// In general, one session should have a Group, whose only one batch events activated at a time.
	// But a EventKey should belong to more than one group, so I used RegisterConfig.StatusMap to solve this problem。
	HandlerMap map[EventKey]EventHandler

	// StatusMap defines an event key is active or inactive in a specific StatusKey.
	// The first EventKey comes from ws.Client to ws.Session, represent client change status.
	// The second EventKey comes from subscribe.Subscribe to Dispatcher, represent data change notification.
	//
	// In a ws.Session, we always register all event handler could be used, but in a period of time,
	// maybe just a few event must be turned on and others must be off.
	StatusMap map[EventKey][]EventKey

	// Session is mainly used to receive and send between ws.Server and ws.Client.
	// Note that Session's ownership in here is borrower from ws.Server, not has it.
	Session ws.Session

	// GetEventKey decodes bytes into a string key and a bytes type data from ws.Client.
	GetEventKey func(req []byte) EventKey
}

var _ innerRegister = &register{}
var _ Register = &register{}

type register struct {
	config        RegisterConfig
	statusMap     map[EventKey]map[EventKey]void
	currStatusKey EventKey
}

func NewRegister(config RegisterConfig) Register {
	var m = make(map[EventKey]map[EventKey]void)
	for k, ks := range config.StatusMap {
		km := make(map[EventKey]void)
		for _, key := range ks {
			km[key] = noop
		}
		m[k] = km
	}
	return &register{
		config:    config,
		statusMap: m,
	}
}

func (sr *register) Run() error {
	go func() {
		for newMessage := range sr.config.Session.Receive() {
			newMessageKey := sr.config.GetEventKey(newMessage)
			if _, has := sr.statusMap[newMessageKey]; has {
				log.Debug("session-register: change keys status, %s", newMessageKey)
				sr.currStatusKey = newMessageKey
			}
		}
	}()
	return nil
}

func (sr *register) Close() error {
	sr.config.Session = nil
	log.Debug("session-register: close")
	return nil
}

func (sr *register) Dispatch(key EventKey, message *[]byte) error {
	// load sr.currStatusKey operation is atomic.
	// fast skip.
	if _, has := sr.statusMap[sr.currStatusKey]; !has {
		return nil
	}

	handler, has := sr.config.HandlerMap[key]
	if !has {
		return fmt.Errorf("resgiter-dispatch: key %s doesn't register", key)
	}

	log.Debug("session-dispatch: get message, %v", string(*message))

	res, err := handler(*message)
	if err != nil {
		return err
	}
	err = sr.config.Session.Send(res)

	log.Debug("session-dispatch: send, %v", string(res))
	return err
}
