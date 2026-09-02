package mailer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cherry/api/internal/store"
)

func TestChannelSelection(t *testing.T) {
	inbox := (&Service{}).Channel()
	if inbox != ChannelInbox {
		t.Fatalf("empty config: %s", inbox)
	}
	smtp := (&Service{SMTP: Config{Host: "smtp.example.com"}}).Channel()
	if smtp != ChannelSMTP {
		t.Fatalf("smtp: %s", smtp)
	}
	resend := (&Service{ResendKey: "re_test", SMTP: Config{Host: "smtp.example.com"}}).Channel()
	if resend != ChannelResend {
		t.Fatalf("resend wins over smtp: %s", resend)
	}
}

func TestSendInboxOnly(t *testing.T) {
	mem := store.NewMemory()
	svc := &Service{Store: mem, WebURL: "http://127.0.0.1:43147"}
	delivery, err := svc.Send(context.Background(), Message{
		To:          "ada@cherry.dev",
		Subject:     "kod",
		PlainBody:   "kodun: 123456",
		UserID:      "u1",
		ChallengeID: "c1",
		Purpose:     store.PurposeNewDevice,
	})
	if err != nil {
		t.Fatal(err)
	}
	if delivery.Sent || delivery.Channel != ChannelInbox {
		t.Fatalf("%#v", delivery)
	}
	mail, err := mem.MailByChallenge(context.Background(), "c1")
	if err != nil || mail.Body != "kodun: 123456" {
		t.Fatalf("mailbox: %v %#v", err, mail)
	}
}

func TestSendRequireWithoutTransport(t *testing.T) {
	mem := store.NewMemory()
	svc := &Service{Store: mem, Require: true}
	_, err := svc.Send(context.Background(), Message{To: "ada@cherry.dev", Subject: "x", PlainBody: "y"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSendResend(t *testing.T) {
	mem := store.NewMemory()
	var gotAuth string
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &payload)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"re_1"}`))
	}))
	t.Cleanup(server.Close)

	svc := &Service{
		Store:      mem,
		ResendKey:  "re_secret",
		ResendFrom: "Cherry <mail@cherry.dev>",
		ResendURL:  server.URL,
		HTTPClient: server.Client(),
	}
	delivery, err := svc.Send(context.Background(), Message{
		To:        "ada@cherry.dev",
		Subject:   "Cherry doğrulama kodu",
		PlainBody: "123456",
		HTMLBody:  "<p>123456</p>",
		UserID:    "u1",
		Purpose:   store.PurposeNewDevice,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !delivery.Sent || delivery.Channel != ChannelResend {
		t.Fatalf("%#v", delivery)
	}
	if gotAuth != "Bearer re_secret" {
		t.Fatalf("auth %q", gotAuth)
	}
	if payload["from"] != "Cherry <mail@cherry.dev>" {
		t.Fatalf("from %#v", payload["from"])
	}
}

func TestCodeEmailAndSMTPBytes(t *testing.T) {
	subject, plain, html := CodeEmail("424242", "tok", "http://127.0.0.1:43147")
	if !strings.Contains(plain, "424242") || !strings.Contains(plain, "/?link=tok") {
		t.Fatalf("plain %q", plain)
	}
	if !strings.Contains(html, "424242") || subject == "" {
		t.Fatalf("html/subject")
	}
	raw := string(smtpBytes("Cherry <a@b.c>", Message{
		To:        "ada@cherry.dev",
		Subject:   "kod",
		PlainBody: plain,
		HTMLBody:  html,
	}))
	if !strings.Contains(raw, "multipart/alternative") || !strings.Contains(raw, "424242") {
		t.Fatalf("smtp bytes %s", raw)
	}
}
