package service

import (
	"context"
	"github.com/webitel/web-meeting-backend/infra/engine"
)

type CallService struct {
	cli *engine.Client
}

func NewCallService(cli *engine.Client) *CallService {
	return &CallService{
		cli: cli,
	}
}

func (s *CallService) SetVariables(ctx context.Context, domainId int64, callId string, vars map[string]string) error {
	return s.cli.SetVariables(ctx, domainId, callId, vars)
}
