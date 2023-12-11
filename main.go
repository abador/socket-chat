package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var userCounter = 0
var chat = &Chat{
	rooms: map[string]*Room{},
	users: map[int]*User{},
}

func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(w, r)
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	userCounter++
	// error with ws so kill goroutine
	if err != nil {
		log.Println(err)
		return
	}
	var user User = User{
		Id:         userCounter,
		LastAlive:  time.Now(),
		Connection: conn,
	}

	for {
		messageType, raw, err := conn.ReadMessage()
		// Can't read message
		// TODO: decide if we want to close the connection or return errors to users in some cases
		if err != nil {
			log.Println(err)
			return
		}
		var req Request

		switch messageType {
		/* Leave ping/pong for now
			case websocket.PingMessage:
			log.Println("user sent ping, responding")
			conn.WriteMessage(websocket.PongMessage, nil)

		case websocket.PongMessage:
			log.Println("user alive")
			// TODO: keepalive
		*/
		case websocket.TextMessage:
			err := json.Unmarshal(raw, &req)
			if err != nil {
				log.Println("message unreadable")
				continue
				// TODO: return error
			}
		default:
			// TODO: decide if we want to return errors to users
			log.Println("invalid message type")
			continue
		}

		err = chat.ProcessRequest(&req, &user)
		if err != nil {
			// TODO: decide if we want to return errors to users
			log.Println(err)
			continue
		}
		resMessage := fmt.Sprintf("Request processed for user %v", user.Id)
		if err := conn.WriteMessage(websocket.TextMessage, []byte(resMessage)); err != nil {
			log.Println(err)
			continue
		}
	}

}
