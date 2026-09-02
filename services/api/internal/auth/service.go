package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cherry/api/internal/crypto"
	"github.com/cherry/api/internal/mailer"
	"github.com/cherry/api/internal/store"
	"github.com/pquerna/otp/totp"
)

const (
	NextSession    = "SESSION"
	NextDeviceCode = "DEVICE_CODE"
	NextTOTP       = "TOTP"
)

type Result struct {
	Next         string
	Token        string
	ChallengeID  string
	User         store.User
	EmailSent    bool
	EmailChannel string
}

type Service struct {
	Store  store.Store
	Mailer *mailer.Service
	Pepper string
	WebURL string
	now    func() time.Time
	limit  *limiter
}

func New(st store.Store, mail *mailer.Service, pepper, webURL string) *Service {
	if pepper == "" {
		pepper = "cherry-dev-pepper-change-me"
	}
	return &Service{
		Store:  st,
		Mailer: mail,
		Pepper: pepper,
		WebURL: strings.TrimRight(webURL, "/"),
		now:    func() time.Time { return time.Now().UTC() },
		limit:  newLimiter(),
	}
}

func (s *Service) Register(ctx context.Context, email, password, deviceFP, deviceLabel, ip string) (Result, error) {
	email = normalizeEmail(email)
	if err := validateCreds(email, password); err != nil {
		return Result{}, err
	}
	if !s.limit.allow("reg:"+email+":"+ip, 8, 15*time.Minute) {
		return Result{}, store.ErrLocked
	}
	hash, err := crypto.HashPassword(password)
	if err != nil {
		return Result{}, err
	}
	user, err := s.Store.CreateUser(ctx, store.User{
		Email:         email,
		PasswordHash:  hash,
		WorkspaceKind: store.WorkspacePersonal,
	})
	if err != nil {
		if errors.Is(err, store.ErrExists) {
			return Result{}, store.ErrExists
		}
		return Result{}, err
	}
	return s.startDeviceChallenge(ctx, user, deviceFP, deviceLabel, store.PurposeNewDevice)
}

func (s *Service) Login(ctx context.Context, email, password, deviceFP, deviceLabel, ip string) (Result, error) {
	email = normalizeEmail(email)
	if err := validateCreds(email, password); err != nil {
		return Result{}, err
	}
	if !s.limit.allow("login:"+email+":"+ip, 10, 15*time.Minute) {
		return Result{}, store.ErrLocked
	}
	user, err := s.Store.GetUserByEmail(ctx, email)
	if err != nil {
		return Result{}, store.ErrInvalidCredentials
	}
	if !crypto.CheckPassword(user.PasswordHash, password) {
		return Result{}, store.ErrInvalidCredentials
	}
	fpHash := s.hashFP(deviceFP)
	device, err := s.Store.GetDeviceByFP(ctx, user.ID, fpHash)
	trusted := err == nil && device.Trusted
	if trusted {
		if user.TotpEnabled {
			return s.startTotpOnly(ctx, *user, deviceFP, deviceLabel)
		}
		return s.issueSession(ctx, *user, deviceFP, deviceLabel, true)
	}
	return s.startDeviceChallenge(ctx, *user, deviceFP, deviceLabel, store.PurposeLoginChallenge)
}

func (s *Service) VerifyCode(ctx context.Context, challengeID, code string, trustDevice bool, ip string) (Result, error) {
	if !s.limit.allow("code:"+challengeID+":"+ip, 8, 10*time.Minute) {
		return Result{}, store.ErrLocked
	}
	challenge, err := s.requireOpenChallenge(ctx, challengeID)
	if err != nil {
		return Result{}, err
	}
	if challenge.Attempts >= challenge.MaxAttempts {
		return Result{}, store.ErrLocked
	}
	want := crypto.HashSecret(s.Pepper, "code", challenge.ID, strings.TrimSpace(code))
	if want != challenge.CodeHash {
		challenge.Attempts++
		_ = s.Store.PutChallenge(ctx, *challenge)
		if challenge.Attempts >= challenge.MaxAttempts {
			return Result{}, store.ErrLocked
		}
		return Result{}, store.ErrInvalidCredentials
	}
	challenge.CodeVerified = true
	challenge.TrustDevice = trustDevice
	if err := s.Store.PutChallenge(ctx, *challenge); err != nil {
		return Result{}, err
	}
	user, err := s.Store.GetUserByID(ctx, challenge.UserID)
	if err != nil {
		return Result{}, err
	}
	if user.TotpEnabled {
		return Result{Next: NextTOTP, ChallengeID: challenge.ID, User: *user}, nil
	}
	return s.finishChallenge(ctx, *challenge, *user, trustDevice)
}

func (s *Service) VerifyLink(ctx context.Context, linkToken, deviceFP, deviceLabel, ip string) (Result, error) {
	if !s.limit.allow("link:"+ip, 20, 15*time.Minute) {
		return Result{}, store.ErrLocked
	}
	linkHash := crypto.HashSecret(s.Pepper, "link", strings.TrimSpace(linkToken))
	challenge, err := s.Store.GetChallengeByLinkHash(ctx, linkHash)
	if err != nil {
		return Result{}, store.ErrInvalidCredentials
	}
	if challenge.Consumed || s.now().After(challenge.ExpiresAt) {
		return Result{}, store.ErrExpired
	}
	challenge.CodeVerified = true
	challenge.TrustDevice = true
	challenge.DeviceFPHash = s.hashFP(deviceFP)
	challenge.DeviceLabel = deviceLabel
	if err := s.Store.PutChallenge(ctx, *challenge); err != nil {
		return Result{}, err
	}
	user, err := s.Store.GetUserByID(ctx, challenge.UserID)
	if err != nil {
		return Result{}, err
	}
	if user.TotpEnabled {
		return Result{Next: NextTOTP, ChallengeID: challenge.ID, User: *user}, nil
	}
	return s.finishChallenge(ctx, *challenge, *user, true)
}

func (s *Service) VerifyTotp(ctx context.Context, challengeID, code, ip string) (Result, error) {
	if !s.limit.allow("totp:"+challengeID+":"+ip, 8, 10*time.Minute) {
		return Result{}, store.ErrLocked
	}
	challenge, err := s.requireOpenChallenge(ctx, challengeID)
	if err != nil {
		return Result{}, err
	}
	if !challenge.CodeVerified {
		return Result{}, store.ErrUnauthorized
	}
	user, err := s.Store.GetUserByID(ctx, challenge.UserID)
	if err != nil {
		return Result{}, err
	}
	if !user.TotpEnabled || user.TotpSecret == "" {
		return Result{}, store.ErrValidation
	}
	if !totp.Validate(strings.TrimSpace(code), user.TotpSecret) {
		return Result{}, store.ErrInvalidCredentials
	}
	return s.finishChallenge(ctx, *challenge, *user, challenge.TrustDevice)
}

func (s *Service) EnableTotp(ctx context.Context, userID string) (secret, otpauth string, err error) {
	user, err := s.Store.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Cherry",
		AccountName: user.Email,
	})
	if err != nil {
		return "", "", err
	}
	user.TotpSecret = key.Secret()
	user.TotpEnabled = false
	if err := s.Store.UpdateUser(ctx, *user); err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

func (s *Service) ConfirmTotp(ctx context.Context, userID, code string) error {
	user, err := s.Store.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.TotpSecret == "" || !totp.Validate(strings.TrimSpace(code), user.TotpSecret) {
		return store.ErrInvalidCredentials
	}
	user.TotpEnabled = true
	return s.Store.UpdateUser(ctx, *user)
}

func (s *Service) DisableTotp(ctx context.Context, userID, code string) error {
	user, err := s.Store.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if !user.TotpEnabled || !totp.Validate(strings.TrimSpace(code), user.TotpSecret) {
		return store.ErrInvalidCredentials
	}
	user.TotpEnabled = false
	user.TotpSecret = ""
	return s.Store.UpdateUser(ctx, *user)
}

func (s *Service) SessionUser(ctx context.Context, token string) (*store.User, *store.Session, error) {
	if token == "" {
		return nil, nil, store.ErrUnauthorized
	}
	sess, err := s.Store.GetSessionByTokenHash(ctx, crypto.HashSecret(s.Pepper, "sess", token))
	if err != nil {
		return nil, nil, store.ErrUnauthorized
	}
	user, err := s.Store.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return nil, nil, store.ErrUnauthorized
	}
	return user, sess, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	sess, err := s.Store.GetSessionByTokenHash(ctx, crypto.HashSecret(s.Pepper, "sess", token))
	if err != nil {
		return nil
	}
	return s.Store.RevokeSession(ctx, sess.ID)
}

func (s *Service) startDeviceChallenge(ctx context.Context, user store.User, deviceFP, deviceLabel string, purpose store.Purpose) (Result, error) {
	if err := s.Store.InvalidateChallenges(ctx, user.ID, purpose); err != nil {
		return Result{}, err
	}
	code, err := crypto.RandomDigits()
	if err != nil {
		return Result{}, err
	}
	link, err := crypto.RandomToken()
	if err != nil {
		return Result{}, err
	}
	challenge := store.Challenge{
		ID:           store.NewID(),
		UserID:       user.ID,
		Purpose:      purpose,
		LinkHash:     crypto.HashSecret(s.Pepper, "link", link),
		Attempts:     0,
		MaxAttempts:  5,
		ExpiresAt:    s.now().Add(10 * time.Minute),
		DeviceFPHash: s.hashFP(deviceFP),
		DeviceLabel:  deviceLabel,
	}
	challenge.CodeHash = crypto.HashSecret(s.Pepper, "code", challenge.ID, code)
	if err := s.Store.PutChallenge(ctx, challenge); err != nil {
		return Result{}, err
	}
	subject, plain, html := mailer.CodeEmail(code, link, s.WebURL)
	delivery, err := s.Mailer.Send(ctx, mailer.Message{
		To:          user.Email,
		Subject:     subject,
		PlainBody:   plain,
		HTMLBody:    html,
		UserID:      user.ID,
		ChallengeID: challenge.ID,
		Purpose:     purpose,
	})
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", store.ErrMailFailed, err)
	}
	return Result{
		Next:         NextDeviceCode,
		ChallengeID:  challenge.ID,
		User:         user,
		EmailSent:    delivery.Sent,
		EmailChannel: delivery.Channel,
	}, nil
}

func (s *Service) startTotpOnly(ctx context.Context, user store.User, deviceFP, deviceLabel string) (Result, error) {
	challenge := store.Challenge{
		ID:           store.NewID(),
		UserID:       user.ID,
		Purpose:      store.PurposeLoginChallenge,
		MaxAttempts:  5,
		ExpiresAt:    s.now().Add(10 * time.Minute),
		DeviceFPHash: s.hashFP(deviceFP),
		DeviceLabel:  deviceLabel,
		CodeVerified: true,
		TrustDevice:  true,
	}
	if err := s.Store.PutChallenge(ctx, challenge); err != nil {
		return Result{}, err
	}
	return Result{Next: NextTOTP, ChallengeID: challenge.ID, User: user}, nil
}

func (s *Service) finishChallenge(ctx context.Context, challenge store.Challenge, user store.User, trustDevice bool) (Result, error) {
	challenge.Consumed = true
	if err := s.Store.PutChallenge(ctx, challenge); err != nil {
		return Result{}, err
	}
	return s.issueSession(ctx, user, "", challenge.DeviceLabel, trustDevice, challenge.DeviceFPHash)
}

func (s *Service) issueSession(ctx context.Context, user store.User, deviceFP, deviceLabel string, trusted bool, fpHash ...string) (Result, error) {
	hash := ""
	if len(fpHash) > 0 && fpHash[0] != "" {
		hash = fpHash[0]
	} else {
		hash = s.hashFP(deviceFP)
	}
	device, err := s.Store.UpsertDevice(ctx, store.Device{
		UserID:   user.ID,
		FPHash:   hash,
		Label:    deviceLabel,
		Trusted:  trusted,
		LastSeen: s.now(),
	})
	if err != nil {
		return Result{}, err
	}
	plain, err := crypto.RandomToken()
	if err != nil {
		return Result{}, err
	}
	sess := store.Session{
		ID:          store.NewID(),
		UserID:      user.ID,
		TokenHash:   crypto.HashSecret(s.Pepper, "sess", plain),
		DeviceID:    device.ID,
		DeviceLabel: device.Label,
		CreatedAt:   s.now(),
	}
	if err := s.Store.CreateSession(ctx, sess); err != nil {
		return Result{}, err
	}
	if err := s.Store.RevokeOtherSessions(ctx, user.ID, sess.ID); err != nil {
		return Result{}, err
	}
	return Result{Next: NextSession, Token: plain, User: user}, nil
}

func (s *Service) requireOpenChallenge(ctx context.Context, id string) (*store.Challenge, error) {
	challenge, err := s.Store.GetChallenge(ctx, id)
	if err != nil {
		return nil, store.ErrInvalidCredentials
	}
	if challenge.Consumed || s.now().After(challenge.ExpiresAt) {
		return nil, store.ErrExpired
	}
	return challenge, nil
}

func (s *Service) hashFP(raw string) string {
	return crypto.HashSecret(s.Pepper, "fp", strings.TrimSpace(raw))
}

func validateCreds(email, password string) error {
	if email == "" || !strings.Contains(email, "@") || strings.TrimSpace(password) == "" {
		return store.ErrValidation
	}
	if len(password) < 8 {
		return store.ErrValidation
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func PurposeLabel(purpose store.Purpose) (string, error) {
	switch purpose {
	case store.PurposeNewDevice:
		return "Yeni cihaz", nil
	case store.PurposeLoginChallenge:
		return "Giriş doğrulama", nil
	case store.PurposeEmailVerify:
		return "E-posta doğrulama", nil
	case store.PurposeSuspiciousLogin:
		return "Şüpheli giriş", nil
	default:
		return "", fmt.Errorf("unhandled purpose: %s", purpose)
	}
}

type limiter struct {
	mu    sync.Mutex
	hits  map[string][]time.Time
}

func newLimiter() *limiter {
	return &limiter{hits: make(map[string][]time.Time)}
}

func (l *limiter) allow(key string, max int, window time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	cut := now.Add(-window)
	kept := make([]time.Time, 0)
	for _, t := range l.hits[key] {
		if t.After(cut) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= max {
		l.hits[key] = kept
		return false
	}
	l.hits[key] = append(kept, now)
	return true
}
