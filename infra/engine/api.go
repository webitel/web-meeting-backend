package engine

import (
	"context"
	"github.com/webitel/web-meeting-backend/gen/engine"
	"github.com/webitel/web-meeting-backend/infra/grpc_client"
	"github.com/webitel/wlog"
)

const ServiceName = "engine"

type Client struct {
	api *grpc_client.Client[engine.CallServiceClient]
	log *wlog.Logger
}

func NewClient(consulTarget string, l *wlog.Logger) (*Client, error) {
	api, err := grpc_client.NewClient[engine.CallServiceClient](consulTarget, ServiceName, engine.NewCallServiceClient)
	if err != nil {
		return nil, err
	}

	return &Client{
		api: api,
		log: l,
	}, nil
}

func (c *Client) SetVariables(ctx context.Context, domainId int64, callId string, vars map[string]string) error {
	_, err := c.api.API.SetVariablesCallNA(ctx, &engine.SetVariablesCallRequestNA{
		Id:        callId,
		DomainId:  domainId,
		Variables: vars,
	})

	return err
}

func (c *Client) Close() error {
	return c.api.Close()
}
