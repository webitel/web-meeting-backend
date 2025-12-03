package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/webitel/web-meeting-backend/internal/model"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/webitel/web-meeting-backend/infra/encrypter"
	"github.com/webitel/wlog"
)

type MeetingStore interface {
	Create(ctx context.Context, m *model.Meeting) error
	Get(ctx context.Context, id string) (*model.Meeting, error)
	Delete(ctx context.Context, id string) error
}

type MeetingService struct {
	ctx       context.Context
	log       *wlog.Logger
	store     MeetingStore
	chat      *ChatService
	encrypter *encrypter.DataEncrypter
}

func NewMeetingService(ctx context.Context, cs *ChatService, log *wlog.Logger, st MeetingStore, enc *encrypter.DataEncrypter) *MeetingService {
	return &MeetingService{
		ctx:       ctx,
		log:       log,
		store:     st,
		encrypter: enc,
		chat:      cs,
	}
}

func (s *MeetingService) CreateMeeting(ctx context.Context, title string, expireSec int64, basePath string, vars map[string]string) (string, string, error) {
	uuid, err := gonanoid.New()
	if err != nil {
		return "", "", err
	}

	now := time.Now().Unix()
	expiresAt := now + expireSec
	if expireSec <= 0 {
		expiresAt = now + 86400 // 24 hours default
	}

	encryptedUuid, err := s.encrypter.Encrypt([]byte(uuid))
	if err != nil {
		return "", "", fmt.Errorf("failed to encrypt meeting id: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(encryptedUuid)

	url := fmt.Sprintf("%s/%s", basePath, token)

	meeting := &model.Meeting{
		Id:        uuid,
		Title:     title,
		CreatedAt: now,
		ExpiresAt: expiresAt,
		Variables: vars,
		Url:       url,
	}

	if err := s.store.Create(ctx, meeting); err != nil {
		return "", "", err
	}

	return token, url, nil
}

func (s *MeetingService) GetMeeting(ctx context.Context, token string) (*model.Meeting, error) {
	// 1. Decode Token
	encryptedUuid, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		// Invalid base64 -> Invalid token -> Don't touch DB
		s.log.Error("invalid token", wlog.Err(err))
		return nil, fmt.Errorf("invalid token")
	}

	// 2. Decrypt Token to get id
	idBytes, err := s.encrypter.Decrypt(encryptedUuid)
	if err != nil {
		// Decryption failed -> Invalid key/token -> Don't touch DB
		s.log.Error("invalid token", wlog.Err(err))
		return nil, fmt.Errorf("invalid token")
	}
	id := string(idBytes)

	// 3. Get from DB using UUID
	meeting, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if meeting == nil {
		return nil, nil // Not found in DB
	}

	// 4. Check expiration
	if time.Now().Unix() > meeting.ExpiresAt {
		return nil, fmt.Errorf("meeting expired")
	}

	return meeting, nil
}

func (s *MeetingService) DeleteMeeting(ctx context.Context, token string) error {
	// Decode & Decrypt first
	encryptedUuid, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return fmt.Errorf("invalid token format: %w", err)
	}
	uuidBytes, err := s.encrypter.Decrypt(encryptedUuid)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}
	uuid := string(uuidBytes)

	return s.store.Delete(ctx, uuid)
}

func (s *MeetingService) CloseChatByMeetingId(ctx context.Context, meetingId string) error {
	return errors.New("TODO")
}
