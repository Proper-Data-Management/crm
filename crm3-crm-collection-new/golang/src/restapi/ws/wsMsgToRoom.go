// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/ext/orm"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils"

	"git.dar.kz/crediton-3/crm-mfo/src/restapi/utils/luautils"
	lua "github.com/Shopify/go-lua"
)

// connection is an middleman between the websocket connection and the hub.
type connectionRoom struct {
	// The websocket connection.
	ws   *websocket.Conn
	room string
	// Buffered channel of outbound messages.
	send chan []byte
}

type tObjMapMsgToRoom struct {
	Data interface{} `json:"data"`
	Room string      `json:"room"`
}

func (c *connectionRoom) RunScript(input interface{}) error {

	o := orm.NewOrm()
	o.Using("default")
	defer o.Rollback()
	o.Begin()
	l := lua.NewState()
	lua.OpenLibraries(l)

	var requestMap = make(map[string]interface{})
	requestMap["input"] = input
	requestMap["room"] = c.room

	luautils.DeepPush(l, requestMap)
	l.SetGlobal("request")

	//loadLuas(l)

	err := luautils.RegisterAPI(l, o)
	luautils.RegisterBPMLUaAPI(nil, l, o)
	if err != nil {
		log.Println("ERROR REGISTER API " + err.Error())
		o.Rollback()
		return err
	}
	luautils.RegisterBPMLUaAPI(nil, l, o)

	log.Println("Startttttt2222")

	var scripts []string
	o.Raw(utils.DbBindReplace("select script from websocket_scripts")).QueryRows(&scripts)

	for _, script := range scripts {
		if err := lua.DoString(l, script); err != nil {
			log.Println("RunLuaServiceScript error lua  " + err.Error())
			//log.Println("RunLuaServiceScript script = "+script)
			//utils.ErrorWriteUser("BPMLuaScriptError","nothing",request.UserId, err)
			l = nil
			o.Rollback()
			return err
		}
	}
	if os.Getenv("CRM_VERBOSE_BPMS") == "1" {
		log.Println("LUA SCRIPT POINT DONE")
	}

	l = nil
	o.Commit()
	return nil
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connectionRoom) msgToRoomReadPump(room_uuid string) {

	o := orm.NewOrm()
	o.Using("default")

	log.Println("room_uuid")
	log.Println(room_uuid)

	var objMap = tObjMapMsgToRoom{}
	defer func() {
		hMsgToRoom.unregister <- c
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

		message, _ = json.Marshal(objMap)

		log.Println("read")
		log.Println(string(message))
		log.Println("room_uuid")
		log.Println(room_uuid)
		log.Println("roomUUid")
		log.Println(room_uuid)

		//o.Raw("insert into bi_chat_logs (user_id,text,created_at,to_account_id) values (?,?,now(),?)",user_id,objMap.Text,objMap.ToAccount).Exec()

		hMsgToRoom.broadcast <- message

		c.RunScript(objMap.Data)

		//log.Println("broadcast")
		//log.Println(objMap.Widget)
		//hMsgToRoom.broadcast <- message

	}
}

// write writes a message with the given message type and payload.
func (c *connectionRoom) msgToRoomWrite(mt int, payload []byte) error {

	//log.Println("server msgToRoomWrite")
	//log.Println(payload)

	c.ws.SetWriteDeadline(time.Now().Add(writeWait))

	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *connectionRoom) msgToRoomWritePump(room string) {

	log.Println("room", room, "c.room", c.room)

	//log.Println("server msgToRoomWritePump")
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	i := 0
	for {
		select {

		case message, ok := <-c.send:
			i++
			log.Println("message22", string(message))
			log.Println("message22", c.room, i, room)
			if !ok {
				c.msgToRoomWrite(websocket.CloseMessage, []byte{})
				return
			}

			//log.Println("before "+string(message))

			var objmap = tObjMapMsgToRoom{}
			err := json.Unmarshal(message, &objmap)
			if err != nil {
				log.Println("ERROR ", err)
			}

			//
			log.Println("after ", objmap, "\r\n", "message:", string(message), "\r\n", "c.room"+c.room, "\r\n", "room:"+objmap.Room)

			if c.room == objmap.Room {
				log.Println("write "+string(message), c.room)
				if err := c.msgToRoomWrite(websocket.TextMessage, message); err != nil {
					return
				}
			}
		case <-ticker.C:
			if err := c.msgToRoomWrite(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeMsgToRoom(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

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
	c := &connectionRoom{send: make(chan []byte, 256), ws: ws, room: p.ByName("room")}
	hMsgToRoom.register <- c
	go c.msgToRoomWritePump(p.ByName("room"))
	c.msgToRoomReadPump(p.ByName("room"))
}
