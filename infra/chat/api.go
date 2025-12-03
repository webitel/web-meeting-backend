package chat

import (
	"context"
	"github.com/webitel/web-meeting-backend/gen/chat"
	"github.com/webitel/web-meeting-backend/infra/grpc_client"
	"github.com/webitel/wlog"
)

const ChatServiceName = "webitel.chat.server"

type Client struct {
	api *grpc_client.Client[chat.ChatServiceClient]
	log *wlog.Logger
}

func NewClient(consulTarget string, l *wlog.Logger) (*Client, error) {
	api, err := grpc_client.NewClient[chat.ChatServiceClient](consulTarget, ChatServiceName, chat.NewChatServiceClient)
	if err != nil {
		return nil, err
	}

	return &Client{
		api: api,
		log: l,
	}, nil
}

func (c *Client) CloseChat(ctx context.Context, convId, closerId string) error {
	_, err := c.api.API.CloseConversation(ctx, &chat.CloseConversationRequest{
		ConversationId:  convId,
		CloserChannelId: closerId,
		Cause:           0,
		AuthUserId:      0,
	})
	return err
}

func (c *Client) Close() error {
	return c.api.Close()
}
