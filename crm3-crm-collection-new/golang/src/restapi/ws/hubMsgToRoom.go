// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ws


func init(){
	go  hMsgToRoom.run()
}
// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hubMsgToRoom struct {
	// Registered connections.
	connections map[*connectionRoom]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *connectionRoom

	// Unregister requests from connections.
	unregister chan *connectionRoom
}

var hMsgToRoom = hubMsgToRoom{
	broadcast:   make(chan []byte),
	register:    make(chan *connectionRoom),
	unregister:  make(chan *connectionRoom),
	connections: make(map[*connectionRoom]bool),
}







func (h *hubMsgToRoom) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}
		case m := <-h.broadcast:
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					close(c.send)
					delete(h.connections, c)
				}
			}
		}
	}
}
