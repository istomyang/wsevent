package wsevent

import (
	"context"
	"encoding/json"
	"github.com/istomyang/wsevent/log"
	"github.com/istomyang/wsevent/subscribe"
	"github.com/istomyang/wsevent/ws"
)

func Example() {

	log.SetStdLogger()
	log.SetLevel(log.DebugLevel)

	// 1. Create a Dispatcher connect with ws.Session and subscribe.Subscribe.
	dis := NewDispatcher(context.Background(), Config{
		MessagePeeler: func(data []byte) (key string, val []byte) {
			var message MessageBody
			_ = json.Unmarshal(data, &message)
			return message.Key, message.Body
		},
	})

	// simulate client send.
	var clientSend chan []byte

	// 2. Create a Ws Server.
	var wsSvr = ws.NewFakeServer(ws.FakeServerConfig{ClientSend: clientSend})
	wsSvr.Run()
	defer wsSvr.Close()

	// 3. Create a ws.Session using ws.Server.
	session1, _ := wsSvr.Create(nil, nil)

	// 4. Create a Register and Register to Dispatcher.
	register1 := NewRegister(RegisterConfig{
		HandlerMap: map[EventKey]EventHandler{
			"event_test1": func(req []byte) ([]byte, error) {
				return []byte("res test_event1"), nil
			},
		},
		StatusMap: map[EventKey][]EventKey{
			"event_default_status": {"event_test1"},
		},
		Session: session1,
		GetEventKey: func(req []byte) EventKey {
			var requestBody ClientRequestBody
			_ = json.Unmarshal(req, &requestBody)
			return requestBody.Key
		},
	})
	_ = dis.Register(register1)

	// simulate publish send.
	var publishSend chan []byte

	// 5. Create subscribe.Subscribe and Connect to Dispatcher.
	sub := subscribe.NewFakeSubscriber(subscribe.FakeConfig{PublishSend: publishSend})
	_ = dis.Connect(sub)

	// 6. Run wsevent
	_ = dis.Run()
	defer dis.Close()

	select {}

	// Output:
}
