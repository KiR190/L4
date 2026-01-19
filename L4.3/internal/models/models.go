package models

import "time"

type Channel string

const (
	Native   Channel = "native"
	Email    Channel = "email"
	Telegram Channel = "telegram"
)

type Notification struct {
	Channel   Channel `json:"channel"`
	Recipient string  `json:"recipient"`
	Message   string  `json:"message"`
}

type ReminderTask struct {
	Notification Notification `json:"notification"`
	SendAt       time.Time    `json:"send_at"`
}

type Event struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	Date            time.Time  `json:"event_date"`
	Title           string     `json:"title"`
	IsArchived      bool       `json:"is_archived"`
	RemindAt        *time.Time `json:"remind_at,omitempty"`
	RemindChannel   Channel    `json:"remind_channel,omitempty"`
	RemindRecipient string     `json:"remind_recipient,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
