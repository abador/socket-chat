package main

import (
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"practice-run/domain"
	"sync"
)

type Chat struct {
	rooms map[string]*domain.Room
	users map[uuid.UUID]*domain.User // TODO: to be able cleanup dead connections etc..
	sync.RWMutex
}

func NewChat() *Chat {
	return &Chat{
		rooms: map[string]*domain.Room{},
		users: map[uuid.UUID]*domain.User{},
	}
}

func (c *Chat) createRoom(name string, user *domain.User) (*domain.Room, error) {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.rooms[name]; !ok {
		c.rooms[name] = domain.NewRoom(name, user)
	}
	return c.rooms[name], nil
}

func (c *Chat) sendMessage(data *domain.Data, user *domain.User) error {
	resp, err := domain.NewMessage(data.Message, user.Id)
	if err != nil {
		return err
	}
	c.RLock()
	defer c.RUnlock()
	if room, ok := c.rooms[data.Room]; ok {
		room.MessageQueue <- resp
		return nil
	}
	return errors.New(fmt.Sprintf("Room %v does not exist", data.Room))
}

func (c *Chat) getRoom(roomName string) (*domain.Room, error) {
	c.RLock()
	defer c.RUnlock()
	room, ok := c.rooms[roomName]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Room %v does not exist", roomName))
	}
	return room, nil
}

func (c *Chat) ProcessRequest(request *domain.Request, user *domain.User) error {
	if request == nil {
		return errors.New("Request is nil")
	}
	if user == nil {
		return errors.New("User is nil")
	}
	spew.Dump(request)
	switch request.Action {
	case domain.ACTION_SUBSCRIBE:
		room, err := c.getRoom(request.Data.Room)
		if err != nil {
			user.SendResponse(domain.RESP_ERROR, "Room does not exist", nil)
			return err
		}
		room.Subscribe(user)
		user.SendResponse(domain.RESP_SUCCESS, "Subscribed to room "+room.Name, nil)
		break
	case domain.ACTION_UNSUBSCRIBE:
		room, err := c.getRoom(request.Data.Room)
		if err != nil {
			user.SendResponse(domain.RESP_ERROR, "Room does not exist", nil)
			return err
		}
		room.Unsubscribe(user)
		user.SendResponse(domain.RESP_SUCCESS, "Unsubscribed from room "+room.Name, nil)
		break
	case domain.ACTION_SEND:
		err := c.sendMessage(&request.Data, user)
		if err != nil {
			return err
		}
		user.SendResponse(domain.RESP_SUCCESS, "Message sent", nil)
		break
	case domain.ACTION_CREATE_ROOM:
		_, origError := c.createRoom(request.Data.Room, user)
		if origError != nil {
			user.SendResponse(domain.RESP_ERROR, "Error with room creation", origError)
		}
		user.SendResponse(domain.RESP_SUCCESS, "Room created", nil)
		break
	default:
		return errors.New("Unsupported action type")
	}
	return nil
}
