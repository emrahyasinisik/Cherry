package mailer

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/icerde/api/internal/store"
)

const (
	ChannelInbox  = "inbox"
	ChannelSMTP   = "smtp"
	ChannelResend = "resend"
)

type Message struct {
	To          string
	Subject     string
	PlainBody   string
	HTMLBody    string
	UserID      string
	ChallengeID string
	Purpose     store.Purpose
}

type Delivery struct {
	Channel string
	Sent    bool
}

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

type Service struct {
	Store      store.Store
	SMTP       Config
	ResendKey  string
	ResendFrom string
	ResendURL  string
	WebURL     string
	Require    bool
	HTTPClient *http.Client
}

func (s *Service) Channel() string {
	if strings.TrimSpace(s.ResendKey) != "" {
		return ChannelResend
	}
	if strings.TrimSpace(s.SMTP.Host) != "" {
		return ChannelSMTP
	}
	return ChannelInbox
}

func (s *Service) Send(ctx context.Context, msg Message) (Delivery, error) {
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
		return Delivery{}, err
	}

	channel := s.Channel()
	switch channel {
	case ChannelInbox:
		if s.Require {
			return Delivery{Channel: channel}, fmt.Errorf("e-posta çıkışı yok: SMTP_HOST veya RESEND_API_KEY tanımla")
		}
		log.Printf("mailer: inbox only (no SMTP/Resend) to=%s", msg.To)
		return Delivery{Channel: ChannelInbox, Sent: false}, nil
	case ChannelSMTP:
		if err := s.sendSMTP(msg); err != nil {
			return Delivery{Channel: channel, Sent: false}, fmt.Errorf("smtp: %w", err)
		}
		log.Printf("mailer: smtp sent to=%s", msg.To)
		return Delivery{Channel: ChannelSMTP, Sent: true}, nil
	case ChannelResend:
		if err := s.sendResend(ctx, msg); err != nil {
			return Delivery{Channel: channel, Sent: false}, fmt.Errorf("resend: %w", err)
		}
		log.Printf("mailer: resend sent to=%s", msg.To)
		return Delivery{Channel: ChannelResend, Sent: true}, nil
	default:
		return Delivery{}, fmt.Errorf("unhandled mail channel: %s", channel)
	}
}

func (s *Service) sendSMTP(msg Message) error {
	port := s.SMTP.Port
	if port == "" {
		port = "587"
	}
	from := s.fromAddress()
	addr := net.JoinHostPort(s.SMTP.Host, port)
	raw := smtpBytes(from, msg)

	dialer := net.Dialer{Timeout: 12 * time.Second}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.SMTP.Host)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: s.SMTP.Host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if strings.TrimSpace(s.SMTP.User) != "" {
		auth := smtp.PlainAuth("", s.SMTP.User, s.SMTP.Password, s.SMTP.Host)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(msg.To); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(raw); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func (s *Service) sendResend(ctx context.Context, msg Message) error {
	from := s.ResendFrom
	if from == "" {
		from = s.fromAddress()
	}
	payload, err := json.Marshal(map[string]any{
		"from":    from,
		"to":      []string{msg.To},
		"subject": msg.Subject,
		"text":    msg.PlainBody,
		"html":    msg.HTMLBody,
	})
	if err != nil {
		return err
	}
	endpoint := s.ResendURL
	if endpoint == "" {
		endpoint = "https://api.resend.com/emails"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.ResendKey)
	req.Header.Set("Content-Type", "application/json")
	client := s.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
	if res.StatusCode >= 300 {
		return fmt.Errorf("status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func (s *Service) fromAddress() string {
	if strings.TrimSpace(s.SMTP.From) != "" {
		return s.SMTP.From
	}
	if strings.TrimSpace(s.ResendFrom) != "" {
		return s.ResendFrom
	}
	return "İçerde <icerde@localhost>"
}

func smtpBytes(from string, msg Message) []byte {
	from = headerSafe(from)
	to := headerSafe(msg.To)
	subject := headerSafe(msg.Subject)
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	if strings.TrimSpace(msg.HTMLBody) == "" {
		b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		b.WriteString(msg.PlainBody)
		return []byte(b.String())
	}
	const boundary = "icerde-alt-7a3f"
	b.WriteString("Content-Type: multipart/alternative; boundary=" + boundary + "\r\n\r\n")
	b.WriteString("--" + boundary + "\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	b.WriteString(msg.PlainBody)
	b.WriteString("\r\n--" + boundary + "\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
	b.WriteString(msg.HTMLBody)
	b.WriteString("\r\n--" + boundary + "--\r\n")
	return []byte(b.String())
}

func headerSafe(value string) string {
	return strings.Map(func(r rune) rune {
		if r == '\r' || r == '\n' {
			return -1
		}
		return r
	}, value)
}

func CodeEmail(code, link, webURL string) (subject, plain, html string) {
	subject = "İçerde doğrulama kodu"
	plain = "İçerde\n\n6 haneli kodun: " + code + "\n10 dakika geçerli, en fazla 5 deneme.\n\nLink: " + webURL + "/?link=" + link + "\n\nSMS kullanılmaz.\n"
	html = `<div style="font-family:IBM Plex Sans,Segoe UI,sans-serif;background:#0E1114;color:#E8E4DC;padding:32px">
<p style="font-size:20px;margin:0 0 12px">İçerde</p>
<p style="color:#8B939C;margin:0 0 20px">6 haneli doğrulama kodun:</p>
<p style="font-family:IBM Plex Mono,monospace;font-size:28px;letter-spacing:8px;color:#C4A574;margin:0 0 20px">` + code + `</p>
<p style="color:#8B939C;font-size:13px">10 dakika geçerli, en fazla 5 deneme.</p>
<p><a href="` + webURL + `/?link=` + link + `" style="color:#C4A574">Link ile doğrula</a></p>
</div>`
	return subject, plain, html
}
