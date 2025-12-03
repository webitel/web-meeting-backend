package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/webitel/web-meeting-backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/webitel/web-meeting-backend/infra/sql"
	"github.com/webitel/wlog"
	"go.uber.org/fx"
)

var Module = fx.Module("store",
	fx.Provide(NewMeetingStore),
)

type MeetingStoreImpl struct {
	log *wlog.Logger
	db  sql.Store
}

func NewMeetingStore(ctx context.Context, db sql.Store, log *wlog.Logger) *MeetingStoreImpl {
	ms := &MeetingStoreImpl{
		log: log,
		db:  db,
	}
	go ms.cleanup(ctx)
	return ms
}

func (s *MeetingStoreImpl) Create(ctx context.Context, m *model.Meeting) error {
	err := s.db.Exec(ctx, `
		INSERT INTO meetings.web_meetings (id, title, created_at, expires_at, variables, url)
		VALUES (@id, @title, @created_at, @expires_at, @variables, @url)
	`, pgx.NamedArgs{
		"id":         m.Id,
		"title":      m.Title,
		"created_at": m.CreatedAt,
		"expires_at": m.ExpiresAt,
		"variables":  m.Variables,
		"url":        m.Url,
	})

	if err != nil {
		return fmt.Errorf("failed to create meeting: %w", err)
	}

	return nil
}

func (s *MeetingStoreImpl) Get(ctx context.Context, id string) (*model.Meeting, error) {
	var m model.Meeting

	err := s.db.Get(ctx, &m, `
		SELECT id, title, created_at, expires_at, variables, url
		FROM meetings.web_meetings
		WHERE id = @id
	`, pgx.NamedArgs{"id": id})

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get meeting: %w", err)
	}

	return &m, nil
}

func (s *MeetingStoreImpl) Delete(ctx context.Context, id string) error {
	err := s.db.Exec(ctx, `DELETE FROM meetings.web_meetings WHERE id = @id`, pgx.NamedArgs{"id": id})
	if err != nil {
		return fmt.Errorf("failed to delete meeting: %w", err)
	}
	return nil
}

func (s *MeetingStoreImpl) DeleteExpires(ctx context.Context, now int64) error {
	err := s.db.Exec(ctx, `DELETE FROM meetings.web_meetings WHERE expires_at <= @now`, pgx.NamedArgs{"now": now})
	if err != nil {
		return fmt.Errorf("failed to delete meeting: %w", err)
	}
	return nil
}

func (s *MeetingStoreImpl) cleanup(ctx context.Context) error {
	timer := time.NewTicker(1 * time.Minute)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			now := time.Now().Unix()
			err := s.DeleteExpires(ctx, now)
			if err != nil {
				s.log.Error("failed to delete expired meetings", wlog.Err(err))
			}
		}
	}
}
