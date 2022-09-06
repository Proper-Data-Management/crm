// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"
	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

type tObjMapMsgToUser struct {
	Data       interface{} `json:"data"`
	Type       string      `json:"type"`
	Room       string      `json:"room"`
	Sender     int64       `json:"sender"`
	SenderName string      `json:"sender_name"`
	Receiver   int64       `json:"receiver"`
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) msgToUserReadPump(user_id int64) {

	o := orm.NewOrm()
	o.Using("default")

	//log.Println("user_id")
	//log.Println(user_id)

	var objMap = tObjMapMsgToUser{}
	defer func() {
		hMsgToUser.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}

		err = json.Unmarshal(message, &objMap)

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}

		//message,_ =json.Marshal(objmap)
		//json.Unmarshal(message, &objmap)

		o.Raw(utils.DbBindReplace("select title from users where id=?"), user_id).QueryRow(&objMap.SenderName)

		objMap.Sender = user_id

		//		if objMap.Receiver == 0 {
		//			objMap.Receiver = user_id
		//		}

		message, _ = json.Marshal(objMap)

		log.Println("read")
		log.Println(string(message))
		log.Println("room")
		log.Println(objMap.Room)

		//o.Raw("insert into bi_chat_logs (user_id,text,created_at,to_account_id) values (?,?,now(),?)",user_id,objMap.Text,objMap.ToAccount).Exec()

		//		if objMap.Room != c.room {
		//			h.broadcast <- message
		//		}

		log.Println("broadcast")
		log.Println(objMap.Receiver)
		log.Println(user_id)
		hMsgToUser.broadcast <- message

	}
}

// write writes a message with the given message type and payload.
func (c *connection) msgToUserWrite(mt int, payload []byte) error {

	//log.Println("server msgToUserWrite")
	//log.Println(payload)

	c.ws.SetWriteDeadline(time.Now().Add(writeWait))

	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *connection) msgToUserWritePump(userId int64) {

	//log.Println("server msgToUserWritePump")
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.msgToUserWrite(websocket.CloseMessage, []byte{})
				return
			}

			//log.Println("before "+string(message))

			var objmap = tObjMapMsgToUser{}
			json.Unmarshal(message, &objmap)
			//			objmap.Sender = userId
			//			message,_ =json.Marshal(objmap)
			//
			//			log.Println("after "+string(message))

			if objmap.Receiver == userId {
				if err := c.msgToUserWrite(websocket.TextMessage, message); err != nil {
					return
				}
			}
		case <-ticker.C:
			if err := c.msgToUserWrite(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeMsgToUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	w.Header().Set("Access-Control-Allow-Origin", "*")

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	r.ParseForm()

	//log.Println("connected...")
	//room:r.Form.Get("room")
	c := &connection{send: make(chan []byte, 256), ws: ws, sender: utils.UserId(r)}
	hMsgToUser.register <- c
	go c.msgToUserWritePump(utils.UserId(r))

	//log.Println("user_id:")
	//log.Println(auth.UserId(r))
	c.msgToUserReadPump(utils.UserId(r))
}
