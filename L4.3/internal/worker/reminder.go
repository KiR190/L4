package worker

import (
	"context"
	"encoding/json"

	"task-manager/internal/logger"
	"task-manager/internal/models"
)

// ReminderWorker обрабатывает задачи напоминаний из канала
type ReminderWorker struct {
	reminderCh <-chan models.ReminderTask
	publisher  Publisher
	stopCh     chan struct{}
}

// Publisher интерфейс для публикации сообщений в очередь
type Publisher interface {
	Publish(body []byte, sendAt interface{}) error
}

// NewReminderWorker создаёт новый воркер
func NewReminderWorker(reminderCh <-chan models.ReminderTask, publisher Publisher) *ReminderWorker {
	return &ReminderWorker{
		reminderCh: reminderCh,
		publisher:  publisher,
		stopCh:     make(chan struct{}),
	}
}

// Start запускает воркер в отдельной горутине
func (w *ReminderWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

func (w *ReminderWorker) run(ctx context.Context) {
	logger.Info("Reminder worker started")
	for {
		select {
		case <-ctx.Done():
			logger.Info("Reminder worker stopping...")
			return
		case <-w.stopCh:
			logger.Info("Reminder worker stopped by Stop()")
			return
		case task := <-w.reminderCh:
			w.processTask(task)
		}
	}
}

func (w *ReminderWorker) processTask(task models.ReminderTask) {
	body, err := json.Marshal(task.Notification)
	if err != nil {
		logger.Error("Failed to marshal notification: %v", err)
		return
	}

	if err := w.publisher.Publish(body, task.SendAt); err != nil {
		logger.Error("Failed to publish notification: %v", err)
		return
	}

	logger.Info("Reminder scheduled for %v: %s", task.SendAt, task.Notification.Message)
}

// Stop останавливает воркер
func (w *ReminderWorker) Stop() {
	close(w.stopCh)
}
