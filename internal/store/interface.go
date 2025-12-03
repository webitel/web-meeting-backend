package store

import (
	"context"
	"github.com/webitel/web-meeting-backend/internal/model"
)

type MeetingStore interface {
	Create(ctx context.Context, m *model.Meeting) error
	Get(ctx context.Context, id string) (*model.Meeting, error)
	Delete(ctx context.Context, id string) error
}
