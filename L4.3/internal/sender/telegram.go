package sender

import (
	"fmt"
	"sync"

	"task-manager/internal/logger"
	"task-manager/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramSender хранит бота и зарегистрированных пользователей
type TelegramSender struct {
	bot   *tgbotapi.BotAPI
	users map[string]int64
	lock  sync.RWMutex
}

// NewTelegramSender создаёт нового бота
func NewTelegramSender(token string) (*TelegramSender, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &TelegramSender{
		bot:   bot,
		users: make(map[string]int64),
	}, nil
}

// ListenAndServe обрабатывает команды пользователей
func (t *TelegramSender) ListenAndServe() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			t.handleCommand(update.Message)
		}
	}
}

// handleCommand регистрирует пользователя по команде /register
func (t *TelegramSender) handleCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "register", "start":
		username := msg.From.UserName
		chatID := msg.Chat.ID

		t.lock.Lock()
		t.users[username] = chatID
		t.lock.Unlock()

		reply := tgbotapi.NewMessage(chatID, "✅ Вы успешно зарегистрированы!\nВаш ID: "+msg.From.UserName)
		_, err := t.bot.Send(reply)
		if err != nil {
			logger.Println("Failed to send registration confirmation:", err)
		}
	default:
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Неизвестная команда")
		_, _ = t.bot.Send(reply)
	}
}

// SendMessageToUser отправляет сообщение зарегистрированному пользователю по username
func (t *TelegramSender) SendMessageToUser(username, text string) error {
	t.lock.RLock()
	chatID, ok := t.users[username]
	t.lock.RUnlock()

	if !ok {
		return fmt.Errorf("user %q not registered, cannot send message", username)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	_, err := t.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send Telegram message to %q: %w", username, err)
	}

	return nil
}

func (t *TelegramSender) Send(notification models.Notification) error {
	return t.SendMessageToUser(notification.Recipient, notification.Message)
}

func (t *TelegramSender) Stop() {
	logger.Println("Stopping Telegram bot...")

	t.bot.StopReceivingUpdates()

	logger.Println("Telegram bot stopped.")
}
