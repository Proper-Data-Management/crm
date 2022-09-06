package ws

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 20 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4096 * 8
)

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws     *websocket.Conn
	sender int64
	// Buffered channel of outbound messages.
	send chan []byte
}
