package publish

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
	"wsevent/merge"
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

func Example_merge() {
	var pub = NewKafkaPublisher(context.Background(), KafkaConfig{
		Hosts: []string{"kafka.service.net"},
		Topic: "topic-test",
	})
	_ = pub.Run()
	defer pub.Close()

	var m = merge.NewMerge(context.Background(), time.Millisecond*500)
	m.Run()
	defer m.Close()

	for i := 0; i < 1000000; i++ {
		var randNumber = rand.Intn(50)

		var k = fmt.Sprint("domain_system_module_componentA_run_fail", randNumber)

		if !m.Allowed(merge.Key(k)) {
			continue
		}

		var message = Message{
			EventName: k,
			Body:      []byte("error: "),
		}
		body, _ := json.Marshal(message)
		_ = pub.Send(body)
	}

	// Output:
}
