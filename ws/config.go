package ws

import (
	"github.com/gorilla/websocket"
)

type ServerConfig struct {
	Upgrader websocket.Upgrader
}

type ClientConfig struct {
}
