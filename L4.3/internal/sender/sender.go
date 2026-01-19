package sender

import (
	"fmt"

	"task-manager/internal/models"
)

type Sender interface {
	Send(notification models.Notification) error
}

type MultiSender struct {
	native   Sender
	email    Sender
	telegram Sender
}

func NewMultiSender(native Sender, email Sender, telegram Sender) *MultiSender {
	return &MultiSender{
		native:   native,
		email:    email,
		telegram: telegram,
	}
}

func (m *MultiSender) Send(notification models.Notification) error {
	switch notification.Channel {
	case models.Native:
		return m.native.Send(notification)
	case models.Email:
		return m.email.Send(notification)
	case models.Telegram:
		return m.telegram.Send(notification)
	default:
		return fmt.Errorf("unknown channel: %v", notification.Channel)
	}
}
