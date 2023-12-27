package domain

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"sync"
	"time"
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
	room := &Room{
		Name:         name,
		Users:        users,
		History:      []*Message{},
		MessageQueue: make(chan *Message, 10),
	}
	go room.listen()
	return room
}

func (r *Room) listen() {
	messageBatch := make([]*Message, 10)
	for {
		select {
		case message := <-r.MessageQueue:
			fmt.Printf("Send message: %v \n", message)
			messageBatch = append(messageBatch, message)
		default:
			if err := r.sendMessages(messageBatch); err != nil {
				fmt.Printf("Problem during batch processing: %v \n", err)
			}
			messageBatch = messageBatch[:0]
			fmt.Println("Sleep 1s")
			time.Sleep(time.Second)
		}
	}
}

func (r *Room) sendMessages(messageBatch []*Message) error {
	for _, message := range messageBatch {
		r.Lock()
		r.History = append(r.History, message)
		r.Unlock()
		mess, err := json.Marshal(message)
		if err != nil {
			return err
		}
		r.RLock()
		fmt.Printf("%v \n", r.Users)
		for _, user := range r.Users {
			fmt.Println(user)
			if err := user.Connection.WriteMessage(websocket.TextMessage, []byte(mess)); err != nil {
				fmt.Printf("Error sending message to user %v", user.Id)
				continue
			}
		}
		r.RUnlock()
	}
	return nil
}

func (r *Room) Subscribe(user *User) error {
	r.Lock()
	if _, ok := r.Users[user.Id]; !ok {
		r.Users[user.Id] = user
	}
	r.Unlock()
	if err := user.SendResponse(RESP_SUCCESS, fmt.Sprintf("Subscribing to  %v", r.Name), nil); err != nil {
		return err
	}
	return nil
}

func (r *Room) Unsubscribe(user *User) error {
	r.Lock()
	if _, ok := r.Users[user.Id]; ok {
		delete(r.Users, user.Id)
	}
	r.Unlock()
	if err := user.SendResponse(RESP_SUCCESS, fmt.Sprintf("Unsubscribing from  %v", r.Name), nil); err != nil {
		return err
	}
	return nil
}
