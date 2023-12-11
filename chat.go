package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/websocket"
	"math/rand"
	"sync"
	"time"
)

const (
	ACTION_SUBSCRIBE = iota
	ACTION_UNSUBSCRIBE
	ACTION_SEND
	ACTION_CREATE_ROOM
	/*
		TODO: for future development
		ACTION_DELETE_ROOM
		ACTION_LOGIN
		ACTION_LOGOUT
	*/
)
const (
	RESP_SUCCESS = iota
	RESP_ERROR
)

type Data struct {
	Room    string `json:"room"`
	Message string `json:"message"`
}

type Request struct {
	Action int  `json:"action"`
	Data   Data `json:"data"`
}

type Response struct {
	Type      int
	Message   string
	OrigError string
}

type Room struct {
	Name    string
	History []*Message
	Users   []*User // for now list since user id might not be unique
	sync.RWMutex
}

func (r *Room) sendMessage(message *Message) error {
	r.RLock()
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

func (r *Room) subscribe(user *User) error {
	r.RLock()
	for _, u := range r.Users {
		if u.Id == user.Id {
			// already subscriber
			return nil
		}
	}
	r.RUnlock()
	r.Lock()
	r.Users = append(r.Users, user)
	r.Unlock()
	user.sendResponse(RESP_SUCCESS, fmt.Sprintf("Subscribing to  %v", r.Name), nil)
	return nil
}

func (r *Room) unsubscribe(user *User) error {
	r.RLock()
	var key int = -1
	for k, u := range r.Users {
		if u.Id == user.Id {
			key = k
			// not a subscriber
			break
		}
	}
	r.RUnlock()
	if key >= 0 {
		r.Lock()
		r.Users = append(r.Users[:key], r.Users[key+1:]...)
		r.Unlock()
	}
	user.sendResponse(RESP_SUCCESS, fmt.Sprintf("Unsubscribing  %v", r.Name), nil)
	return nil
}

type Message struct {
	Id      int
	Sent    time.Time
	Message string
	UserId  int
}

type User struct {
	Id         int
	LastAlive  time.Time
	Connection *websocket.Conn
	sync.RWMutex
}

func (u *User) sendResponse(t int, message string, origError error) error {
	var resp = Response{
		Type:    t,
		Message: message,
		// TODO: issue with originating error
	}
	resMessage, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	if err := u.Connection.WriteMessage(websocket.TextMessage, []byte(resMessage)); err != nil {
		return err
	}
	return nil
}

type Chat struct {
	rooms map[string]*Room
	users map[int]*User // TODO: to be able cleanup dead connections etc..
	sync.RWMutex
}

func (c *Chat) CreateRoom(name string, user *User) (*Room, error) {
	c.RLock()
	var room *Room = nil
	room, ok := c.rooms[name]
	c.RUnlock()
	if !ok {
		c.Lock()
		room = &Room{
			Name:    name,
			Users:   []*User{user},
			History: []*Message{},
		}
		c.rooms[name] = room
		c.Unlock()
	}
	return room, nil
}

func (c *Chat) SendMessage(data *Data, user *User) error {
	var resp = Message{
		Message: data.Message,
		Id:      rand.Int(), // this is just for the PoC, not focusing on the identifier
		UserId:  user.Id,
		Sent:    time.Now(),
	}
	c.RLock()
	room, ok := c.rooms[data.Room]
	c.RUnlock()
	if !ok {
		return errors.New(fmt.Sprintf("Room %v does not exist", data.Room))
	}
	return room.sendMessage(&resp)
}

func (c *Chat) subsccribe(data *Data, user *User) error {
	var resp = Message{
		Message: data.Message,
		Id:      rand.Int(), // this is just for the PoC, not focusing on the identifier
		UserId:  user.Id,
		Sent:    time.Now(),
	}
	c.RLock()
	room, ok := c.rooms[data.Room]
	c.RUnlock()
	if !ok {
		return errors.New(fmt.Sprintf("Room %v does not exist", data.Room))
	}
	return room.sendMessage(&resp)
}

func (c *Chat) getRoom(roomName string) (*Room, error) {
	c.RLock()
	defer c.RUnlock()
	room, ok := c.rooms[roomName]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Room %v does not exist", roomName))
	}
	return room, nil
}

func (c *Chat) ProcessRequest(request *Request, user *User) error {
	if request == nil {
		return errors.New("Request is nil")
	}
	if user == nil {
		return errors.New("User is nil")
	}
	spew.Dump(request)
	switch request.Action {
	case ACTION_SUBSCRIBE:
		room, err := c.getRoom(request.Data.Room)
		if err != nil {
			user.sendResponse(RESP_ERROR, "Room does not exist", nil)
		}
		room.subscribe(user)
		user.sendResponse(RESP_SUCCESS, "Subscribed to room "+room.Name, nil)
		break
	case ACTION_UNSUBSCRIBE:
		room, err := c.getRoom(request.Data.Room)
		if err != nil {
			user.sendResponse(RESP_ERROR, "Room does not exist", nil)
		}
		room.unsubscribe(user)
		user.sendResponse(RESP_SUCCESS, "Subscribed to room "+room.Name, nil)
		break
	case ACTION_SEND:
		err := c.SendMessage(&request.Data, user)
		if err != nil {
			return err
		}
		user.sendResponse(RESP_SUCCESS, "Message sent", nil)
		break
	case ACTION_CREATE_ROOM:
		_, origError := c.CreateRoom(request.Data.Room, user)
		if origError != nil {
			user.sendResponse(RESP_ERROR, "Error with room creation", origError)
		}
		user.sendResponse(RESP_SUCCESS, "Room created", nil)
		break
	default:
		return errors.New("Unsupported action type")
	}
	return nil
}
