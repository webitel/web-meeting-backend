package handler

import (
	"context"

	api "github.com/webitel/web-meeting-backend/gen/go/api/meetings"
)

type MeetingHandler interface {
}

type MeetingService struct {
	api.UnimplementedMeetingServiceServer
}

func (m *MeetingService) CreateMeeting(ctx context.Context, request *api.CreateMeetingRequest) (*api.Meeting, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MeetingService) GetMeeting(ctx context.Context, request *api.GetMeetingRequest) (*api.Meeting, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MeetingService) GetMeetingNA(ctx context.Context, request *api.GetMeetingRequest) (*api.Meeting, error) {
	//TODO implement me
	panic("implement me")
}
