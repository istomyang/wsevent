package subscribe

import (
	"context"
	"encoding/json"
)

type Message struct {
	EventName string
	Body      []byte
}

func Example_ws() {
	var sub = NewWsSubscriber(context.Background(), WsConfig{
		Addr: "localhost:8080",
		Path: "/ws-test",
	})
	_ = sub.Run()
	defer sub.Close()

	messageChan, _ := sub.Get()

	for message := range messageChan {
		var m Message
		_ = json.Unmarshal(message, &m)
	}

	// Output:
}

func Example_kafka() {
	var sub = NewKafkaSubscriber(context.Background(), KafkaConfig{
		Hosts:       []string{"localhost:3096"},
		PartitionID: 0,
		Topic:       "topic-test",
	})
	_ = sub.Run()
	defer sub.Close()

	messageChan, _ := sub.Get()

	for message := range messageChan {
		var m Message
		_ = json.Unmarshal(message, &m)
	}

	// Output:
}
