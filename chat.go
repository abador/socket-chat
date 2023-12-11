package main

import (
	"encoding/json"
	"errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

const (
	ACTION_SUBSCRIBE = iota
	ACTION_UNSUBSCRIBE
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
	Room string `json:"room"`
}

type Request struct {
	Action int `json:"action"`
	Data   `json:"data"`
}

type Response struct {
	Type      int
	Message   string
	OrigError string
}

type Room struct {
	Name  string
	Users []*User
	sync.RWMutex
}

type User struct {
	Id         int
	LastAlive  time.Time
	Connection *websocket.Conn
	sync.RWMutex
}

type Chat struct {
	rooms map[string]*Room
	users map[int]*User
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
			Name:  name,
			Users: []*User{user},
		}
		c.rooms[name] = room
		c.Unlock()
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
		break
	case ACTION_UNSUBSCRIBE:
		break
	case ACTION_CREATE_ROOM:
		_, origError := c.CreateRoom(request.Room, user)
		if origError != nil {
			var resp = Response{
				Type:      RESP_ERROR,
				Message:   "Error with room creation",
				OrigError: origError.Error(),
			}
			resMessage, err := json.Marshal(resp)
			if err != nil {
				return errors.Join(origError, err)
			}
			if err := user.Connection.WriteMessage(websocket.TextMessage, []byte(resMessage)); err != nil {
				return errors.Join(origError, err)
			}
		}
		var resp = Response{
			Type:    RESP_SUCCESS,
			Message: "Message sent",
		}
		resMessage, err := json.Marshal(resp)
		if err != nil {
			return err
		}
		if err := user.Connection.WriteMessage(websocket.TextMessage, []byte(resMessage)); err != nil {
			return errors.Join(origError, err)
		}
		break
	default:
		return errors.New("Unsupported action type")
	}
	return nil
}
