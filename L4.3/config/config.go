package config

import (
	"fmt"

	"github.com/DeanPDX/dotconfig"
)

type Config struct {
	AppPort          string `env:"APP_PORT" default:"8080"`

	// Database Settings
	DatabaseHost     string `env:"DATABASE_HOST" default:"localhost"`
	DatabasePort     int    `env:"DATABASE_PORT" default:"5432"`
	DatabaseUser     string `env:"DATABASE_USER" default:"calendar_user"`
	DatabasePassword string `env:"DATABASE_PASSWORD" default:"mysecretpassword"`
	DatabaseName     string `env:"DATABASE_NAME" default:"calendar_db"`
	DatabaseSSLMode  string `env:"DATABASE_SSLMODE" default:"disable"`

	// RabbitMQ
	RabbitURL        string `env:"RABBIT_URL" default:"amqp://guest:guest@rabbitmq:5672/"`
	ReminderQueueCap int    `env:"REMINDER_QUEUE_CAP" default:"256"`

	// SMTP Settings
	SMTPHost     string `env:"SMTP_HOST" default:"smtp.gmail.com"`
	SMTPPort     int    `env:"SMTP_PORT" default:"587"`
	SMTPUsername string `env:"SMTP_USERNAME"`
	SMTPPassword string `env:"SMTP_PASSWORD"`
	SMTPFrom     string `env:"SMTP_FROM"`

	// Telegram Bot
	TelegramBotToken string `env:"TG_BOT_TOKEN"`

	// Архивирование
	ArchiveIntervalMinutes int `env:"ARCHIVE_INTERVAL_MINUTES" default:"60"`
	ArchiveCutoffDays      int `env:"ARCHIVE_CUTOFF_DAYS" default:"30"`
}

func Load() (Config, error) {
	cfg, err := dotconfig.FromFileName[Config](".env")
	if err != nil {
		fmt.Printf("Error loading config: %v. Using defaults...\n", err)
	}

	fmt.Printf("Loaded config: %+v\n", cfg)

	return cfg, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DatabaseUser, c.DatabasePassword, c.DatabaseHost, c.DatabasePort, c.DatabaseName, c.DatabaseSSLMode)
}
