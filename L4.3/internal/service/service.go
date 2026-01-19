package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"task-manager/internal/models"
	"task-manager/internal/repo"
)

type Service struct {
	repo       repo.EventRepo
	reminderCh chan<- models.ReminderTask
}

func NewService(r repo.EventRepo, reminderCh chan<- models.ReminderTask) *Service {
	return &Service{repo: r, reminderCh: reminderCh}
}

func (s *Service) CreateEvent(ctx context.Context, e *models.Event) error {
	if err := s.repo.CreateEvent(ctx, e); err != nil {
		return err
	}

	return s.scheduleReminder(e)
}

func (s *Service) UpdateEvent(ctx context.Context, e *models.Event) error {
	return s.repo.UpdateEvent(ctx, e)
}

func (s *Service) DeleteEvent(ctx context.Context, id int64) error {
	return s.repo.DeleteEvent(ctx, id)
}

func (s *Service) GetEventsForDay(ctx context.Context, userID int64, date time.Time) ([]models.Event, error) {
	return s.repo.GetEventsForDay(ctx, userID, date)
}

func (s *Service) GetEventsForRange(ctx context.Context, userID int64, start, end time.Time) ([]models.Event, error) {
	return s.repo.GetEventsForRange(ctx, userID, start, end)
}

func (s *Service) scheduleReminder(e *models.Event) error {
	if s.reminderCh == nil || e.RemindAt == nil || e.RemindAt.IsZero() {
		return nil
	}

	channel := e.RemindChannel
	if channel == "" {
		channel = models.Native
	}
	if channel != models.Native && e.RemindRecipient == "" {
		return errors.New("remind_recipient is required for non-native channel")
	}

	task := models.ReminderTask{
		Notification: models.Notification{
			Channel:   channel,
			Recipient: e.RemindRecipient,
			Message:   e.Title,
		},
		SendAt: *e.RemindAt,
	}

	select {
	case s.reminderCh <- task:
		return nil
	default:
		return fmt.Errorf("reminder queue is full")
	}
}
