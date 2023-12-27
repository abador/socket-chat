package domain

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type User struct {
	Id         uuid.UUID
	LastAlive  time.Time
	Connection *websocket.Conn
	sync.RWMutex
}

func NewUser(conn *websocket.Conn) (*User, error) {
	userId, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	return &User{
		Id:         userId,
		LastAlive:  time.Now(),
		Connection: conn,
	}, nil
}

func (u *User) SendResponse(t int, message string, origError error) error {
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
