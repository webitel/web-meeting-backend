package service

import (
	"context"
	"github.com/webitel/web-meeting-backend/infra/chat"
	"github.com/webitel/wlog"
)

type ChatService struct {
	log *wlog.Logger
	api *chat.Client
}

func NewChatService(log *wlog.Logger, api *chat.Client) *ChatService {
	return &ChatService{
		log: log,
		api: api,
	}
}

func (c *ChatService) CloseChat(ctx context.Context, conversationId, closerId string, authUserId int64) error {

	return c.api.CloseChat(ctx, conversationId, closerId, authUserId)
}
