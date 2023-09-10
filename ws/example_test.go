package ws

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/istomyang/wsevent/log"
	"net/http"
	"sync"
	"time"
)

func Example() {
	log.SetStdLogger()

	go runSvr()

	// Wait for server initialized.
	t := time.After(time.Second * 3)
	<-t

	go runClient()

	select {
	case <-time.After(time.Second * 30):
		return
	}

	// Output:
}

func runSvr() {

	// Create a server.
	var svr = NewServer(context.Background(), ServerConfig{
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	})
	svr.Run()
	defer svr.Close()

	// Simulate a pub/sub connection.
	var otherSource = newFakeSource()
	otherSource.Run()
	defer otherSource.Close()

	// You can also use other http framework having http.ResponseWriter and *http.Request.
	http.HandleFunc("/ws-test", func(w http.ResponseWriter, r *http.Request) {

		// Get Session.
		session, _ := svr.Create(w, r)

		var wg sync.WaitGroup

		// You can handle messages from ws client.
		// In general, client send one message, and server send many messages.
		// You can serve multiple inputs as ingress, aggregate those and make an output as egress.
		wg.Add(1)
		go func() {
			defer wg.Done()
			for req := range session.Receive() {
				var data = []byte(fmt.Sprintf("server-send-res: %s", string(req)))
				_ = session.Send(data)
			}
		}()

		// You can handle your business logic in this block, and push to session if needed.
		wg.Add(1)
		go func() {
			defer wg.Done()
			for number := range otherSource.Get() {
				var data = []byte(fmt.Sprintf("server-send-sub: %d", number))
				_ = session.Send(data)
			}
		}()

		wg.Wait()
	})

	log.Info("Server listening at:", "8081")
	_ = http.ListenAndServe(":8081", nil)
}

func runClient() {
	var addr = "ws://localhost:8081"
	var path = "/ws-test"

	client := NewClient(context.Background(), ClientConfig{})
	client.Run()
	defer client.Close()

	s, err := client.Create(addr, path)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for data := range s.Receive() {
			log.Info("client-get: ", string(data))
		}
	}()

	// You can run send code in a loop.
	_ = s.Send([]byte("client-send: message1"))
	_ = s.Send([]byte("client-send: message2"))

	wg.Wait()
}

// fakeSource send a fake message to you regularly, simulating pub/sub.
type fakeSource struct {
	data chan int
}

func newFakeSource() fakeSource {
	return fakeSource{data: make(chan int)}
}

func (s *fakeSource) Get() <-chan int {
	return s.data
}

func (s *fakeSource) Run() {
	go func() {
		var number = 0
		for range time.Tick(time.Second * 1) { // Allow leak.
			if number > 3 {
				break
			}
			s.data <- number
			number += 1
		}
	}()
}

func (s *fakeSource) Close() {
	close(s.data)
}
