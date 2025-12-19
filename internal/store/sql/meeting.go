package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/webitel/web-meeting-backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/webitel/web-meeting-backend/infra/sql"
	"github.com/webitel/wlog"
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
	//go ms.cleanup(ctx) // TODO
	return ms
}

func (s *MeetingStoreImpl) Create(ctx context.Context, m *model.Meeting) error {
	err := s.db.Exec(ctx, `
		INSERT INTO meetings.web_meetings (id, domain_id, title, created_at, expires_at, variables, url)
		VALUES (@id, @domain_id, @title, @created_at, @expires_at, @variables, @url)
	`, pgx.NamedArgs{
		"id":         m.Id,
		"domain_id":  m.DomainId,
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
		SELECT id, domain_id, title, created_at, expires_at, variables, url, call_id, satisfaction, bridged
		FROM meetings.web_meetings
		WHERE id = @id
	`, pgx.NamedArgs{"id": id})

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get meeting %s: %w", id, err)
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

func (s *MeetingStoreImpl) SetCall(ctx context.Context, id string, callId string, bridged bool) error {
	err := s.db.Exec(ctx, `update meetings.web_meetings
set call_id = @call_id,
    bridged = @bridged
where id = @id`, pgx.NamedArgs{
		"id":      id,
		"call_id": callId,
		"bridged": bridged,
	})

	if err != nil {
		return fmt.Errorf("failed to set call_id: %w", err)
	}

	return nil
}

func (s *MeetingStoreImpl) SetSatisfaction(ctx context.Context, id string, satisfaction string) error {
	err := s.db.Exec(ctx, `update meetings.web_meetings
set satisfaction = @satisfaction
where id = @id`, pgx.NamedArgs{
		"id":           id,
		"satisfaction": satisfaction,
	})

	if err != nil {
		return fmt.Errorf("failed to set satisfaction: %w", err)
	}

	return nil
}

// TODO move to service
func (s *MeetingStoreImpl) GetChatCloseInfo(ctx context.Context, id string) (*model.ChatCloseInfo, error) {
	var res model.ChatCloseInfo
	err := s.db.Get(ctx, &res, `select c.id as conversation_id, ch.id closer_id, ch.user_id as auth_user_id
from chat.conversation c
    left join chat.channel ch on ch.conversation_id = c.id
where (c.props->>'wbt_meeting_id') = @id
    and c.closed_at isnull
order by c.created_at desc, ch.created_at
limit 1`, pgx.NamedArgs{
		"id": id,
	})

	if err != nil {
		if s.db.IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, err
	}

	return &res, nil
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
