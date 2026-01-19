package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"task-manager/internal/mocks"
	"task-manager/internal/models"
	"task-manager/internal/service"
)

func TestCreateEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepo(ctrl)
	svc := service.NewService(mockRepo, nil)

	e := &models.Event{Title: "Test Event"}

	mockRepo.EXPECT().
		CreateEvent(gomock.Any(), e).
		Return(nil)

	err := svc.CreateEvent(context.Background(), e)
	assert.NoError(t, err)
}

func TestUpdateEvent_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepo(ctrl)
	svc := service.NewService(mockRepo, nil)

	e := &models.Event{ID: 1, Title: "Bad Event"}

	mockRepo.EXPECT().
		UpdateEvent(gomock.Any(), e).
		Return(errors.New("update failed"))

	err := svc.UpdateEvent(context.Background(), e)
	assert.EqualError(t, err, "update failed")
}

func TestDeleteEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepo(ctrl)
	svc := service.NewService(mockRepo, nil)

	mockRepo.EXPECT().
		DeleteEvent(gomock.Any(), int64(42)).
		Return(nil)

	err := svc.DeleteEvent(context.Background(), 42)
	assert.NoError(t, err)
}

func TestGetEventsForDay(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepo(ctrl)
	svc := service.NewService(mockRepo, nil)

	date := time.Now()
	expected := []models.Event{{ID: 1, Title: "Meeting"}}

	mockRepo.EXPECT().
		GetEventsForDay(gomock.Any(), int64(5), date).
		Return(expected, nil)

	events, err := svc.GetEventsForDay(context.Background(), 5, date)
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "Meeting", events[0].Title)
}

func TestGetEventsForRange_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepo(ctrl)
	svc := service.NewService(mockRepo, nil)

	from := time.Now()
	to := from.Add(24 * time.Hour)

	mockRepo.EXPECT().
		GetEventsForRange(gomock.Any(), int64(1), from, to).
		Return(nil, errors.New("db down"))

	events, err := svc.GetEventsForRange(context.Background(), 1, from, to)
	assert.Error(t, err)
	assert.Nil(t, events)
	assert.EqualError(t, err, "db down")
}
