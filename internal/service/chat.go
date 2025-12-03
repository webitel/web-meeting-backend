package service

import (
	"context"
	"github.com/webitel/web-meeting-backend/infra/chat"
	"github.com/webitel/wlog"
)

type ChatService struct {
	ctx context.Context
	log *wlog.Logger
	api *chat.Client
}

func NewChatService(ctx context.Context, log *wlog.Logger, api *chat.Client) *ChatService {
	return &ChatService{
		ctx: ctx,
		log: log,
		api: api,
	}
}

func (c *ChatService) CloseChat(meetingId string) error {
	conversationId := meetingId
	closerId := meetingId

	return c.api.CloseChat(c.ctx, conversationId, closerId)
}
