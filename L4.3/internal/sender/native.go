package sender

import (
	"task-manager/internal/logger"
	"task-manager/internal/models"
)

type NativeSender struct{}

func (n *NativeSender) Send(notification models.Notification) error {
	logger.Printf("📨 Получено сообщение: %+v\n", notification)
	return nil
}
