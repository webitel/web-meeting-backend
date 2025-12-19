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
	SetCall(ctx context.Context, id string, callId string, bridged bool) error
	SetSatisfaction(ctx context.Context, id string, satisfaction string) error

	GetChatCloseInfo(ctx context.Context, id string) (*model.ChatCloseInfo, error)
}

type MeetingService struct {
	ctx       context.Context
	log       *wlog.Logger
	store     MeetingStore
	chat      *ChatService
	call      *CallService
	encrypter *encrypter.DataEncrypter
}

func NewMeetingService(ctx context.Context, cs *ChatService, call *CallService, log *wlog.Logger, st MeetingStore, enc *encrypter.DataEncrypter) *MeetingService {
	return &MeetingService{
		ctx:       ctx,
		log:       log,
		store:     st,
		encrypter: enc,
		chat:      cs,
		call:      call,
	}
}

func (s *MeetingService) CreateMeeting(ctx context.Context, domainId int64, title string, expireSec int64, basePath string, vars map[string]string) (string, string, error) {
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
		DomainId:  domainId,
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

func (s *MeetingService) GetMeeting(ctx context.Context, meetingId string) (*model.Meeting, error) {
	id, err := s.decodeToken(meetingId)
	if err != nil {
		return nil, err
	}

	meeting, err := s.store.Get(ctx, id)
	if err != nil {
		s.log.Error(err.Error(), wlog.Err(err))
		return nil, nil
	}
	if meeting == nil {
		return nil, nil // Not found in DB
	}

	meeting.AllowSatisfaction = meeting.Bridged && meeting.CallId != nil && meeting.Satisfaction == nil

	return meeting, nil
}

func (s *MeetingService) DeleteMeeting(ctx context.Context, meetingId string) error {
	id, err := s.decodeToken(meetingId)
	if err != nil {
		return err
	}

	return s.store.Delete(ctx, id)
}

func (s *MeetingService) decodeToken(meetingId string) (string, error) {
	encryptedUuid, err := base64.URLEncoding.DecodeString(meetingId)
	if err != nil {
		return "", fmt.Errorf("invalid token format: %w", err)
	}
	uuidBytes, err := s.encrypter.Decrypt(encryptedUuid)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	return string(uuidBytes), nil
}

func (s *MeetingService) CloseByCall(ctx context.Context, meetingId string, callId string, bridged bool) (string, error) {
	id, err := s.decodeToken(meetingId)
	if err != nil {
		return "", err
	}

	err = s.store.SetCall(ctx, id, callId, bridged)
	if err != nil {
		return id, err
	}

	chatInfo, err := s.store.GetChatCloseInfo(ctx, meetingId)
	if err != nil {
		s.log.Error(err.Error(), wlog.Err(err))
		return id, nil
	}

	if chatInfo == nil {
		return id, nil
	}

	return id, s.chat.CloseChat(ctx, chatInfo.ConversationId, chatInfo.CloserId, chatInfo.AuthUserId)
}

func (s *MeetingService) Satisfaction(ctx context.Context, meetingId string, satisfaction string) error {
	meeting, err := s.GetMeeting(ctx, meetingId)
	if err != nil {
		return err
	}

	if meeting == nil {
		return errors.New("meeting not found")
	}

	if meeting.CallId == nil || (meeting.Satisfaction != nil && len(*meeting.Satisfaction) > 0) {
		return fmt.Errorf("not allow")
	}

	err = s.call.SetVariables(ctx, meeting.DomainId, *meeting.CallId, map[string]string{
		model.MeetingSatisfactionVarName: satisfaction,
	})

	if err != nil {
		// TODO
		return err
	}

	return s.store.SetSatisfaction(ctx, meeting.Id, satisfaction)
}
