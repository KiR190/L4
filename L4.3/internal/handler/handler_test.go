package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"task-manager/internal/handler"
	"task-manager/internal/models"
)

// --- Mock Service ---
type MockService struct {
	mock.Mock
}

func (m *MockService) CreateEvent(ctx context.Context, e *models.Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *MockService) UpdateEvent(ctx context.Context, e *models.Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *MockService) DeleteEvent(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockService) GetEventsForRange(ctx context.Context, userID int64, from, to time.Time) ([]models.Event, error) {
	args := m.Called(ctx, userID, from, to)
	return args.Get(0).([]models.Event), args.Error(1)
}

// --- Tests ---

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &MockService{}
	h := handler.NewHandler(svc, func(c *gin.Context) { c.Next() })

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest("GET", "/healthz", nil)
	h.Health(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &MockService{}
	h := handler.NewHandler(svc, func(c *gin.Context) { c.Next() })

	event := models.Event{Title: "Test Event"}
	svc.On("CreateEvent", mock.Anything, &event).Return(nil)

	body, _ := json.Marshal(event)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/create_event", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.CreateEvent(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertCalled(t, "CreateEvent", mock.Anything, &event)
}

func TestUpdateEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &MockService{}
	h := handler.NewHandler(svc, func(c *gin.Context) { c.Next() })

	event := models.Event{ID: 1, Title: "Updated Event"}
	svc.On("UpdateEvent", mock.Anything, &event).Return(nil)

	body, _ := json.Marshal(event)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/update_event", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateEvent(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertCalled(t, "UpdateEvent", mock.Anything, &event)
}

func TestDeleteEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &MockService{}
	h := handler.NewHandler(svc, func(c *gin.Context) { c.Next() })

	reqBody := map[string]int64{"id": 1}
	svc.On("DeleteEvent", mock.Anything, int64(1)).Return(nil)

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/delete_event", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.DeleteEvent(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertCalled(t, "DeleteEvent", mock.Anything, int64(1))
}

func TestEventsForDay(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &MockService{}
	h := handler.NewHandler(svc, func(c *gin.Context) { c.Next() })

	userID := int64(1)
	events := []models.Event{{ID: 1, Title: "Event1"}}

	// Используем mock.AnythingOfType("time.Time"), чтобы не падало на timezone
	svc.On(
		"GetEventsForRange",
		mock.Anything,                    // ctx
		userID,                           // userID
		mock.AnythingOfType("time.Time"), // start
		mock.AnythingOfType("time.Time"), // end
	).Return(events, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/events_for_day?user_id=1&date=2025-10-27", nil)
	c.Request = req

	h.EventsForDay(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertCalled(
		t,
		"GetEventsForRange",
		mock.Anything,
		userID,
		mock.AnythingOfType("time.Time"),
		mock.AnythingOfType("time.Time"),
	)
}
