package handler

import (
	"context"
	"errors"
	grpc_srv "github.com/webitel/web-meeting-backend/infra/grpc_srv"
	"github.com/webitel/web-meeting-backend/internal/model"

	wmb "github.com/webitel/web-meeting-backend/gen/web-meeting-backend"
	"github.com/webitel/wlog"
)

type MeetingService interface {
	CreateMeeting(ctx context.Context, title string, expireSec int64, basePath string, vars map[string]string) (string, string, error)
	GetMeeting(ctx context.Context, id string) (*model.Meeting, error)
	DeleteMeeting(ctx context.Context, id string) error
	CloseChatByMeetingId(ctx context.Context, meetingId string) error
}

type MeetingHandler struct {
	log *wlog.Logger
	svc MeetingService
	wmb.UnimplementedMeetingServiceServer
}

func NewMeetingHandler(svc MeetingService, s *grpc_srv.Server, l *wlog.Logger) *MeetingHandler {

	h := &MeetingHandler{
		svc: svc,
		log: l,
	}
	wmb.RegisterMeetingServiceServer(s, h)

	return h
}

func (h *MeetingHandler) CreateMeeting(ctx context.Context, request *wmb.CreateMeetingRequest) (*wmb.CreateMeetingResponse, error) {
	id, url, err := h.svc.CreateMeeting(ctx, request.Title, request.ExpireSec, request.BasePath, request.Variables)
	if err != nil {
		h.log.Error("failed to create meeting", wlog.Err(err))
		return nil, err
	}

	return &wmb.CreateMeetingResponse{
		Id:  id,
		Url: url,
	}, nil
}

func (h *MeetingHandler) GetMeeting(ctx context.Context, request *wmb.GetMeetingRequest) (*wmb.Meeting, error) {
	meeting, err := h.svc.GetMeeting(ctx, request.Id)
	if err != nil {
		h.log.Error("failed to get meeting", wlog.Err(err))
		return nil, err
	}
	if meeting == nil {
		return nil, nil // gRPC поверне OK з nil body, або можна повернути status.NotFound
	}

	return &wmb.Meeting{
		Id:        meeting.Id,
		Title:     meeting.Title,
		CreatedAt: meeting.CreatedAt,
		ExpiresAt: meeting.ExpiresAt,
		Variables: meeting.Variables,
		Url:       meeting.Url,
	}, nil
}

func (h *MeetingHandler) DeleteMeeting(ctx context.Context, request *wmb.DeleteMeetingRequest) (*wmb.DeleteMeetingResponse, error) {
	err := h.svc.DeleteMeeting(ctx, request.Id)
	if err != nil {
		h.log.Error("failed to delete meeting", wlog.Err(err))
		return nil, err
	}
	return &wmb.DeleteMeetingResponse{}, nil
}

func (h *MeetingHandler) SatisfactionMeeting(ctx context.Context, request *wmb.SatisfactionMeetingRequest) (*wmb.SatisfactionMeetingResponse, error) {
	return nil, errors.New("TODO")
}
