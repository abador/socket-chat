package domain

import (
	"github.com/google/uuid"
	"time"
)

type Message struct {
	Id      uuid.UUID
	Sent    time.Time
	Message string
	UserId  uuid.UUID
}

func NewMessage(message string, userId uuid.UUID) (*Message, error) {
	messageId, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	return &Message{
		Message: message,
		Id:      messageId,
		UserId:  userId,
		Sent:    time.Now(),
	}, nil
}
