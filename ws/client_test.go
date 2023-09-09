package ws

import (
	"context"
	"github.com/gorilla/websocket"
	"net/http"
	"testing"
	"time"
	"wsevent/log"
)

func TestMain(m *testing.M) {
	log.SetStdLogger()

	// Run Server
	go func() {
		http.HandleFunc("/ws-test", func(w http.ResponseWriter, r *http.Request) {
			var upgrader = websocket.Upgrader{
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			}

			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Error(err)
				return
			}
			//goland:noinspection ALL
			defer conn.Close()

			_ = conn.WriteMessage(websocket.BinaryMessage, []byte("hello, client."))
			log.Debug("server-send: hello, client.")

			for {
				_, data, err := conn.ReadMessage()
				if err != nil {
					break
				}
				log.Debug("server-get:", string(data))
			}
		})
		log.Debug("Server started on :8081")
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()

	// Wait for server initialized.
	t := time.After(time.Second * 3)
	<-t

	m.Run()
}

func TestClient(t *testing.T) {
	var addr = "ws://localhost:8081"
	var path = "/ws-test"

	client := NewClient(context.Background(), ClientConfig{})
	client.Run()
	defer client.Close()

	go CloseAfter(client, time.Second*20)

	s, err := client.Create(addr, path)
	if err != nil {
		panic(err)
	}

	_ = s.Send([]byte("client-send: hello."))

	for res := range s.Receive() {
		log.Debug("client-subscribe: ", string(res))
	}
}

func CloseAfter(client Client, duration time.Duration) {
	select {
	case <-time.After(duration):
		client.Close()
		return
	}
}
