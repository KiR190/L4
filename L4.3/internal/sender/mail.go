package sender

import (
	"time"

	"task-manager/internal/logger"
	"task-manager/internal/models"

	mail "github.com/xhit/go-simple-mail/v2"
)

// EmailSender хранит настройки SMTP
type EmailSender struct {
	Server   string
	Port     int
	Username string
	Password string
	From     string
}

// NewEmailSender создает новый объект EmailSender
func NewEmailSender(server string, port int, username, password, from string) *EmailSender {
	return &EmailSender{
		Server:   server,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}

// SendEmail отправляет письмо на указанный адрес с заданным текстом
func (s *EmailSender) SendEmail(to, body string) error {
	// Настраиваем SMTP сервер
	smtp := mail.NewSMTPClient()
	smtp.Host = s.Server
	smtp.Port = s.Port
	smtp.Username = s.Username
	smtp.Password = s.Password
	smtp.Encryption = mail.EncryptionSTARTTLS
	smtp.KeepAlive = false
	smtp.ConnectTimeout = 10 * time.Second
	smtp.SendTimeout = 10 * time.Second

	// Подключаемся к SMTP серверу
	client, err := smtp.Connect()
	if err != nil {
		logger.Println("SMTP connection error:", err)
		return err
	}

	// Создаем письмо
	email := mail.NewMSG()
	email.SetFrom(s.From).
		AddTo(to).
		SetSubject("Напоминание").
		SetBody(mail.TextPlain, body)

	// Отправляем письмо
	if err := email.Send(client); err != nil {
		logger.Println("Error sending email:", err)
		return err
	}

	logger.Println("Email sent to", to)
	return nil
}

func (s *EmailSender) Send(notification models.Notification) error {
	return s.SendEmail(notification.Recipient, notification.Message)
}
