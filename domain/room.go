package domain

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"sync"
)

type Room struct {
	Name         string
	History      []*Message
	Users        map[uuid.UUID]*User // for now list since user id might not be unique
	MessageQueue chan *Message
	sync.RWMutex
}

func NewRoom(name string, user *User) *Room {
	users := make(map[uuid.UUID]*User)
	users[user.Id] = user
	return &Room{
		Name:    name,
		Users:   users,
		History: []*Message{},
	}
}

func (r *Room) SendMessage(message *Message) error {
	r.Lock()
	r.History = append(r.History, message)
	r.Unlock()
	mess, err := json.Marshal(message)
	if err != nil {
		return err
	}
	r.RLock()
	for _, user := range r.Users {
		if err := user.Connection.WriteMessage(websocket.TextMessage, []byte(mess)); err != nil {
			fmt.Printf("Error sending message to user %v", user.Id)
			continue
		}
	}
	r.RUnlock()
	return nil
}

func (r *Room) Subscribe(user *User) error {
	r.Lock()
	if _, ok := r.Users[user.Id]; !ok {
		r.Users[user.Id] = user
	}
	r.Unlock()
	user.SendResponse(RESP_SUCCESS, fmt.Sprintf("Subscribing to  %v", r.Name), nil)
	return nil
}

func (r *Room) Unsubscribe(user *User) error {
	r.Lock()
	if _, ok := r.Users[user.Id]; ok {
		delete(r.Users, user.Id)
	}
	r.Unlock()
	user.SendResponse(RESP_SUCCESS, fmt.Sprintf("Unsubscribing  %v", r.Name), nil)
	return nil
}
