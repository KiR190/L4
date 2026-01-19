package worker

import (
	"context"
	"time"

	"task-manager/internal/logger"
)

// ArchiveRepo интерфейс для архивации событий
type ArchiveRepo interface {
	ArchiveOldEvents(ctx context.Context, cutoffDate time.Time) (int64, error)
}

// ArchiveWorker фоновый воркер для архивации старых событий
type ArchiveWorker struct {
	repo     ArchiveRepo
	interval time.Duration
	cutoff   time.Duration
	stopCh   chan struct{}
}

// NewArchiveWorker создаёт новый воркер архивации
func NewArchiveWorker(repo ArchiveRepo, intervalMinutes, cutoffDays int) *ArchiveWorker {
	return &ArchiveWorker{
		repo:     repo,
		interval: time.Duration(intervalMinutes) * time.Minute,
		cutoff:   time.Duration(cutoffDays) * 24 * time.Hour,
		stopCh:   make(chan struct{}),
	}
}

// Start запускает воркер в отдельной горутине
func (w *ArchiveWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

func (w *ArchiveWorker) run(ctx context.Context) {
	logger.Info("Archive worker started (interval: %v, cutoff: %v)", w.interval, w.cutoff)

	// Первый запуск сразу при старте
	w.archiveOldEvents(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Archive worker stopping...")
			return
		case <-w.stopCh:
			logger.Info("Archive worker stopped by Stop()")
			return
		case <-ticker.C:
			w.archiveOldEvents(ctx)
		}
	}
}

func (w *ArchiveWorker) archiveOldEvents(ctx context.Context) {
	cutoffDate := time.Now().Add(-w.cutoff)

	archived, err := w.repo.ArchiveOldEvents(ctx, cutoffDate)
	if err != nil {
		logger.Error("Failed to archive old events: %v", err)
		return
	}

	if archived > 0 {
		logger.Info("Archived %d events older than %s", archived, cutoffDate.Format("2006-01-02"))
	}
}

// Stop останавливает воркер
func (w *ArchiveWorker) Stop() {
	close(w.stopCh)
}
