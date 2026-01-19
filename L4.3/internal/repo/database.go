package repo

import (
	"context"
	"fmt"
	"time"

	"task-manager/internal/models"
)

func (r *Repository) CreateEvent(ctx context.Context, e *models.Event) error {
	query := `INSERT INTO events (user_id, event_date, title, is_archived)
			  VALUES ($1, $2, $3, $4)
			  RETURNING id, created_at, updated_at`
	return r.Pool.QueryRow(ctx, query, e.UserID, e.Date, e.Title, false).
		Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
}

func (r *Repository) UpdateEvent(ctx context.Context, e *models.Event) error {
	query := `UPDATE events SET event_date=$1, title=$2 WHERE id=$3 AND is_archived=false RETURNING created_at, updated_at`
	return r.Pool.QueryRow(ctx, query, e.Date, e.Title, e.ID).Scan(&e.CreatedAt, &e.UpdatedAt)
}

func (r *Repository) DeleteEvent(ctx context.Context, id int64) error {
	cmdTag, err := r.Pool.Exec(ctx, `DELETE FROM events WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("event not found")
	}
	return nil
}

func (r *Repository) GetEventsForDay(ctx context.Context, userID int64, date time.Time) ([]models.Event, error) {
	query := `SELECT id, user_id, event_date, title, is_archived, created_at, updated_at
			  FROM events WHERE user_id=$1 AND event_date=$2 AND is_archived=false`
	return r.fetch(ctx, query, userID, date)
}

func (r *Repository) GetEventsForRange(ctx context.Context, userID int64, from, to time.Time) ([]models.Event, error) {
	query := `SELECT id, user_id, event_date, title, is_archived, created_at, updated_at
			  FROM events WHERE user_id=$1 AND event_date BETWEEN $2 AND $3 AND is_archived=false ORDER BY event_date`
	return r.fetch(ctx, query, userID, from, to)
}

// ArchiveOldEvents помечает как архивные события старше cutoffDate
func (r *Repository) ArchiveOldEvents(ctx context.Context, cutoffDate time.Time) (int64, error) {
	query := `UPDATE events SET is_archived = true WHERE event_date < $1 AND is_archived = false`
	cmdTag, err := r.Pool.Exec(ctx, query, cutoffDate)
	if err != nil {
		return 0, err
	}
	return cmdTag.RowsAffected(), nil
}

func (r *Repository) fetch(ctx context.Context, query string, args ...interface{}) ([]models.Event, error) {
	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(&e.ID, &e.UserID, &e.Date, &e.Title, &e.IsArchived, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
