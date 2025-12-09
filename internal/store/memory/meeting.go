package memory

import (
	"context"
	"errors"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/webitel/web-meeting-backend/internal/model"
	"github.com/webitel/wlog"
	"go.uber.org/fx"
)

const (
	cacheSize = 50000
)

var Module = fx.Module("store",
	fx.Provide(NewMeetingMemoryStore),
)

type MeetingMemoryStore struct {
	log  *wlog.Logger
	data *lru.Cache[string, *model.Meeting]
}

func NewMeetingMemoryStore(log *wlog.Logger) *MeetingMemoryStore {
	cache, _ := lru.New[string, *model.Meeting](cacheSize)
	return &MeetingMemoryStore{
		log:  log,
		data: cache,
	}
}

func (s *MeetingMemoryStore) Create(ctx context.Context, m *model.Meeting) error {
	s.data.Add(m.Id, m)
	return nil
}

func (s *MeetingMemoryStore) Get(ctx context.Context, id string) (*model.Meeting, error) {
	meeting, ok := s.data.Get(id)
	if !ok {
		return nil, errors.New("meeting not found")
	}
	if meeting.ExpiresAt <= time.Now().Unix() {
		s.Delete(ctx, id)
		return nil, errors.New("meeting expired")
	}
	return meeting, nil
}

func (s *MeetingMemoryStore) Delete(ctx context.Context, id string) error {
	ok := s.data.Remove(id)
	if !ok {
		return errors.New("meeting not found")
	}
	return nil
}

func (s *MeetingMemoryStore) SetCallId(ctx context.Context, id string, callId string) error {
	panic("DEL")
}
