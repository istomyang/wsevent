package publish

import (
	"context"
	"encoding/json"
)

type Message struct {
	EventName string
	Body      []byte
}

func Example_kafka() {

	var pub = NewKafkaPublisher(context.Background(), KafkaConfig{
		Hosts: []string{"kafka.service.net"},
		Topic: "topic-test",
	})
	_ = pub.Run()
	defer pub.Close()

	var message = Message{
		EventName: "domain_system_module_componentA_run_fail",
		Body:      []byte("error: "),
	}

	body, _ := json.Marshal(message)

	_ = pub.Send(body)

	// Output:
}
