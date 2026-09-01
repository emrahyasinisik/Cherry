package mailer

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	"github.com/icerde/api/internal/store"
)

type Message struct {
	To          string
	Subject     string
	PlainBody   string
	UserID      string
	ChallengeID string
	Purpose     store.Purpose
}

type Mailer interface {
	Send(ctx context.Context, msg Message) error
}

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

type Service struct {
	Store  store.Store
	SMTP   Config
	WebURL string
}

func (s *Service) Send(ctx context.Context, msg Message) error {
	mail := store.Mail{
		ID:          store.NewID(),
		UserID:      msg.UserID,
		ChallengeID: msg.ChallengeID,
		Subject:     msg.Subject,
		Body:        msg.PlainBody,
		Purpose:     msg.Purpose,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.Store.AddMail(ctx, mail); err != nil {
		return err
	}
	if strings.TrimSpace(s.SMTP.Host) == "" {
		log.Printf("mailer: delivered to in-app mailbox (no SMTP) to=%s subject=%q", msg.To, msg.Subject)
		return nil
	}
	addr := s.SMTP.Host + ":" + s.SMTP.Port
	from := s.SMTP.From
	if from == "" {
		from = "icerde@localhost"
	}
	auth := smtp.PlainAuth("", s.SMTP.User, s.SMTP.Password, s.SMTP.Host)
	raw := []byte("To: " + msg.To + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + msg.Subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		msg.PlainBody)
	if err := smtp.SendMail(addr, auth, from, []string{msg.To}, raw); err != nil {
		return fmt.Errorf("smtp: %w", err)
	}
	return nil
}

func CodeEmail(code, link, webURL string) (subject, body string) {
	subject = "İçerde doğrulama kodu"
	body = "İçerde\n\n6 haneli kodun: " + code + "\n10 dakika geçerli, en fazla 5 deneme.\n\nLink: " + webURL + "/?link=" + link + "\n\nSMS kullanılmaz.\n"
	return subject, body
}
