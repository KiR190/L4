package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"task-manager/config"
	"task-manager/internal/handler"
	"task-manager/internal/logger"
	"task-manager/internal/models"
	"task-manager/internal/queue"
	"task-manager/internal/repo"
	"task-manager/internal/sender"
	"task-manager/internal/service"
	"task-manager/internal/worker"
)

func main() {
	// Загружаем конфиг
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	logger.Info("Starting server on port %s...", cfg.AppPort)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Подключение к БД
	db, err := repo.Connect(cfg.DatabaseURL())
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Проверка соединения
	if err := db.Ping(ctx); err != nil {
		logger.Fatalf("Database not reachable: %v", err)
	}
	logger.Info("Connected to database successfully")

	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Reminder queue and RabbitMQ setup
	reminderCh := make(chan models.ReminderTask, cfg.ReminderQueueCap)

	notifqueue, err := queue.NewQueue(cfg.RabbitURL)
	if err != nil {
		logger.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer func() {
		if closeErr := notifqueue.Close(); closeErr != nil {
			logger.Error("Queue close error: %v", closeErr)
		}
	}()

	// Telegram
	token := cfg.TelegramBotToken
	telegramSender, err := sender.NewTelegramSender(token)
	if err != nil {
		logger.Fatalf("Failed to create Telegram sender: %v", err)
	}
	go telegramSender.ListenAndServe()

	emailSender := sender.NewEmailSender(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.SMTPFrom,
	)

	consoleSender := &sender.NativeSender{}

	multiSender := sender.NewMultiSender(consoleSender, emailSender, telegramSender)

	notifqueue.SetHandler(func(ctx context.Context, n models.Notification) error {
		return multiSender.Send(n)
	})
	notifqueue.StartMainConsumer()
	notifqueue.StartDLQConsumer()

	// Запуск воркера напоминаний
	reminderWorker := worker.NewReminderWorker(reminderCh, notifqueue)
	reminderWorker.Start(shutdownCtx)

	// Запуск воркера архивации
	archiveWorker := worker.NewArchiveWorker(db, cfg.ArchiveIntervalMinutes, cfg.ArchiveCutoffDays)
	archiveWorker.Start(shutdownCtx)

	// Инициализация сервиса
	taskService := service.NewService(db, reminderCh)
	// Получаем middleware для логирования
	requestLoggerMiddleware := logger.RequestLogger()
	// Инициализация handler
	ginHandler := handler.NewHandler(taskService, requestLoggerMiddleware)

	// Запуск HTTP сервера
	if err := ginHandler.Run(shutdownCtx, cfg.AppPort); err != nil {
		logger.Error("Failed to start HTTP server: %v", err)
	}

	// Graceful shutdown
	logger.Info("Shutting down gracefully...")
	logger.Close()
}
