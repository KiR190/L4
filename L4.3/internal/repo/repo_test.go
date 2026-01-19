package repo_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"task-manager/internal/models"
	"task-manager/internal/repo"
)

func setupMock(t *testing.T) (pgxmock.PgxPoolIface, repo.EventRepo) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	return mock, &repo.Repository{Pool: mock}
}

func TestCreateEvent(t *testing.T) {
	mock, r := setupMock(t)
	defer mock.Close()

	e := &models.Event{
		UserID: 1,
		Date:   time.Now(),
		Title:  "Test Event",
	}

	mock.ExpectQuery(`INSERT INTO events`).
		WithArgs(e.UserID, e.Date, e.Title, false).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(10, time.Now(), time.Now()))

	err := r.CreateEvent(context.Background(), e)
	require.NoError(t, err)
	assert.Equal(t, int64(10), e.ID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateEvent(t *testing.T) {
	mock, r := setupMock(t)
	defer mock.Close()

	e := &models.Event{ID: 5, Title: "Updated", Date: time.Now()}

	mock.ExpectQuery(`UPDATE events SET event_date`).
		WithArgs(e.Date, e.Title, e.ID).
		WillReturnRows(pgxmock.NewRows([]string{"created_at", "updated_at"}).
			AddRow(e.CreatedAt, e.UpdatedAt))

	err := r.UpdateEvent(context.Background(), e)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteEvent_Success(t *testing.T) {
	mock, r := setupMock(t)
	defer mock.Close()

	mock.ExpectExec(`DELETE FROM events WHERE id=\$1`).
		WithArgs(int64(7)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := r.DeleteEvent(context.Background(), 7)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteEvent_NotFound(t *testing.T) {
	mock, r := setupMock(t)
	defer mock.Close()

	mock.ExpectExec(`DELETE FROM events WHERE id=\$1`).
		WithArgs(int64(99)).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := r.DeleteEvent(context.Background(), 99)
	require.EqualError(t, err, "event not found")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetEventsForDay(t *testing.T) {
	mock, r := setupMock(t)
	defer mock.Close()

	date := time.Now().Truncate(24 * time.Hour)

	mock.ExpectQuery(`SELECT id, user_id, event_date, title, is_archived, created_at, updated_at`).
		WithArgs(int64(1), date).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "event_date", "title", "is_archived", "created_at", "updated_at",
		}).AddRow(1, 1, date, "Meeting", false, time.Now(), time.Now()))

	events, err := r.GetEventsForDay(context.Background(), 1, date)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "Meeting", events[0].Title)
}

func TestGetEventsForRange_Error(t *testing.T) {
	mock, r := setupMock(t)
	defer mock.Close()

	from, to := time.Now(), time.Now().Add(24*time.Hour)
	mock.ExpectQuery(`SELECT id, user_id, event_date, title`).
		WithArgs(int64(1), from, to).
		WillReturnError(errors.New("db error"))

	events, err := r.GetEventsForRange(context.Background(), 1, from, to)
	require.Error(t, err)
	require.Nil(t, events)
}
