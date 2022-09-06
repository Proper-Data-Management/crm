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

type tObjMap struct {
	Text       string `json:"text"`
	Room       string `json:"room"`
	Sender     int64  `json:"sender"`
	SenderName string `json:"sender_name"`
	ToAccount  int64  `json:"to_account"`
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) accountChatReadPump(user_id int64) {

	o := orm.NewOrm()
	o.Using("default")

	var objMap = tObjMap{}
	defer func() {
		hAccountChat.unregister <- c
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

		if objMap.ToAccount == 0 {
			o.Raw(utils.DbBindReplace("select account_id from users where id=?"), user_id).QueryRow(&objMap.ToAccount)
		}

		message, _ = json.Marshal(objMap)

		log.Println("read")
		log.Println(string(message))

		o.Raw("insert into bi_chat_logs (user_id,text,created_at,to_account_id) values (?,?,now(),?)", user_id, objMap.Text, objMap.ToAccount).Exec()

		//		if objMap.Room != c.room {
		//			h.broadcast <- message
		//		}
		hAccountChat.broadcast <- message
	}
}

// write writes a message with the given message type and payload.
func (c *connection) accountChatWrite(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))

	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *connection) accountChatReadWritePump(userId int64) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.accountChatWrite(websocket.CloseMessage, []byte{})
				return
			}

			//			log.Println("before "+string(message))
			//		    var objmap = Tobjmap{}
			//			json.Unmarshal(message, &objmap)
			//			objmap.Sender = userId
			//			message,_ =json.Marshal(objmap)
			//
			//			log.Println("after "+string(message))

			if err := c.accountChatWrite(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.accountChatWrite(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeAccountChat(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	r.ParseForm()

	//log.Println("chat..."+(r.Form.Get("room")))
	//room:r.Form.Get("room")
	c := &connection{send: make(chan []byte, 256), ws: ws, sender: utils.UserId(r)}
	hAccountChat.register <- c
	go c.accountChatReadWritePump(utils.UserId(r))
	c.accountChatReadPump(utils.UserId(r))
}
