package services

import (
	"fmt"
	"log/slog"
	"net/smtp"
)

type EmailService interface {
	SendWelcomeEmail(to, name string) error
}

type emailService struct {
	host     string
	port     string
	user     string
	from     string
	password string
	log      *slog.Logger
}

func NewEmailService(host, port, from, user, password string, log *slog.Logger) EmailService {
	return &emailService{
		host:     host,
		port:     port,
		user:     user,
		from:     from,
		password: password,
		log:      log,
	}
}

func (s *emailService) SendWelcomeEmail(to, name string) error {
	fromHeader := fmt.Sprintf("From: Fiber Go AppMock <%s>\n", s.from)
	toHeader := fmt.Sprintf("To: %s\n", to)
	subject := "Subject: Welcome to fiber-go-mock!\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf("<h1>Hi, %s</h1><p>Регистрация прошла успешно!</p>", name)
	msg := []byte(fromHeader + toHeader + subject + mime + body)

	auth := smtp.PlainAuth("", s.user, s.password, s.host)

	s.log.Info("", s.user, s.from)

	// Попытка отправить сообщение
	err := smtp.SendMail(s.host+":"+s.port, auth, s.from, []string{to}, msg)
	if err != nil {
		s.log.Error("SMTP Error: не удалось отправить сообщение",
			"to", to,
			"error", err)
		return err
	}

	s.log.Info("Сообщение отправлено успешно!", "to", to)
	return nil

}
