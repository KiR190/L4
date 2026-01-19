package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"task-manager/internal/logger"
	"task-manager/internal/models"
)

type Handler struct {
	service EventService
	router  *gin.Engine
}

// NewHandler создает новый Handler с middlewares
func NewHandler(svc EventService, requestLogger gin.HandlerFunc) *Handler {
	r := gin.New()
	r.Use(gin.Recovery(), requestLogger)

	h := &Handler{
		service: svc,
		router:  r,
	}
	h.RegisterRoutes(r)
	return h
}

type EventService interface {
	CreateEvent(ctx context.Context, e *models.Event) error
	UpdateEvent(ctx context.Context, e *models.Event) error
	DeleteEvent(ctx context.Context, id int64) error
	GetEventsForRange(ctx context.Context, userID int64, from, to time.Time) ([]models.Event, error)
}

// Run запускает HTTP-сервер и корректно завершает работу при отмене ctx
func (h *Handler) Run(ctx context.Context, port string) error {
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: h.router,
	}

	go func() {
		logger.Info("Server started on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %s", err)
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down server gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	logger.Info("Server exited properly")
	return nil
}

// Route Handlers
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// bindJSON — обертка для JSON binding с автоматической ошибкой
func bindJSON[T any](c *gin.Context) (T, bool) {
	var obj T
	if err := c.ShouldBindJSON(&obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		var zero T
		return zero, false
	}
	return obj, true
}

// parseQueryParamInt64 — парсинг int64 query параметра
func parseQueryParamInt64(c *gin.Context, key string) (int64, bool) {
	valStr := c.Query(key)
	val, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid %s", key)})
		return 0, false
	}
	return val, true
}

// parseQueryParamDate — парсинг даты query параметра
func parseQueryParamDate(c *gin.Context, key string) (time.Time, bool) {
	valStr := c.Query(key)
	t, err := time.Parse("2006-01-02", valStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid %s", key)})
		return time.Time{}, false
	}
	return t, true
}

func (h *Handler) CreateEvent(c *gin.Context) {
	e, ok := bindJSON[models.Event](c)
	if !ok {
		return
	}
	if err := h.service.CreateEvent(c, &e); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": e})
}

func (h *Handler) UpdateEvent(c *gin.Context) {
	e, ok := bindJSON[models.Event](c)
	if !ok {
		return
	}
	if err := h.service.UpdateEvent(c, &e); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": e})
}

func (h *Handler) DeleteEvent(c *gin.Context) {
	req, ok := bindJSON[struct {
		ID int64 `json:"id" binding:"required"`
	}](c)
	if !ok {
		return
	}
	if err := h.service.DeleteEvent(c, req.ID); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "deleted"})
}

func (h *Handler) EventsForDay(c *gin.Context) {
	h.fetchEventsByRange(c, "day")
}

func (h *Handler) EventsForWeek(c *gin.Context) {
	h.fetchEventsByRange(c, "week")
}

func (h *Handler) EventsForMonth(c *gin.Context) {
	h.fetchEventsByRange(c, "month")
}

// fetchEventsByRange функция для получения событий
func (h *Handler) fetchEventsByRange(c *gin.Context, period string) {
	userID, ok := parseQueryParamInt64(c, "user_id")
	if !ok {
		return
	}
	date, ok := parseQueryParamDate(c, "date")
	if !ok {
		return
	}

	var start, end time.Time

	switch period {
	case "day":
		start = date
		end = date
	case "week":
		weekday := int(date.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start = date.AddDate(0, 0, -weekday+1)
		end = start.AddDate(0, 0, 6)

	case "month":
		start = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		end = start.AddDate(0, 1, -1)
	}

	/*start := date
	end := start.AddDate(0, months, days)*/

	events, err := h.service.GetEventsForRange(c, userID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": events})
}

// Routes
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/healthz", h.Health)
	r.POST("/create_event", h.CreateEvent)
	r.POST("/update_event", h.UpdateEvent)
	r.POST("/delete_event", h.DeleteEvent)
	r.GET("/events_for_day", h.EventsForDay)
	r.GET("/events_for_week", h.EventsForWeek)
	r.GET("/events_for_month", h.EventsForMonth)
}
