package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/webitel/web-meeting-backend/infra/encrypter"
	"github.com/webitel/web-meeting-backend/internal/model"
	"github.com/webitel/wlog"
)

// MockMeetingStore is a mock implementation of the MeetingStore interface
type MockMeetingStore struct {
	mock.Mock
}

func (m *MockMeetingStore) Create(ctx context.Context, meeting *model.Meeting) error {
	args := m.Called(ctx, meeting)
	return args.Error(0)
}

func (m *MockMeetingStore) Get(ctx context.Context, id string) (*model.Meeting, error) {
	args := m.Called(ctx, id)
	if meeting, ok := args.Get(0).(*model.Meeting); ok {
		return meeting, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockMeetingStore) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMeetingStore) SetCallId(ctx context.Context, id string, callId string) error {
	args := m.Called(ctx, id, callId)
	return args.Error(0)
}

func (m *MockMeetingStore) SetSatisfaction(ctx context.Context, id string, satisfaction string) error {
	args := m.Called(ctx, id, satisfaction)
	return args.Error(0)
}

func setupMeetingService(t *testing.T) (*MeetingService, *MockMeetingStore) {
	mockStore := new(MockMeetingStore)
	logger := wlog.NewLogger(&wlog.LoggerConfiguration{EnableConsole: false})

	// Use a fixed key for reproducible tests
	key := []byte("12345678901234567890123456789012")
	enc, err := encrypter.New(key)
	require.NoError(t, err)

	svc := NewMeetingService(context.Background(), nil, nil, logger, mockStore, enc)
	return svc, mockStore
}

func TestMeetingService_CreateMeeting(t *testing.T) {
	svc, mockStore := setupMeetingService(t)
	ctx := context.Background()

	domainID := int64(1)
	title := "Test Meeting"
	expireSec := int64(3600)
	basePath := "https://example.com/meeting"
	vars := map[string]string{"key": "value"}

	// Expect Create to be called
	mockStore.On("Create", ctx, mock.AnythingOfType("*model.Meeting")).Return(nil).Run(func(args mock.Arguments) {
		meeting := args.Get(1).(*model.Meeting)
		assert.NotEmpty(t, meeting.Id)
		assert.Equal(t, domainID, meeting.DomainId)
		assert.Equal(t, title, meeting.Title)
		assert.Equal(t, vars, meeting.Variables)
		assert.True(t, meeting.ExpiresAt > time.Now().Unix())
		assert.Contains(t, meeting.Url, basePath)
	})

	token, url, err := svc.CreateMeeting(ctx, domainID, title, expireSec, basePath, vars)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, url)
	assert.Contains(t, url, token)

	mockStore.AssertExpectations(t)
}

func TestMeetingService_GetMeeting(t *testing.T) {
	// Helper to get a valid token and ID for testing
	getValidTokenAndID := func(t *testing.T, svc *MeetingService, mockStore *MockMeetingStore) (string, string) {
		ctx := context.Background()
		var generatedID string
		// We temporarily mock Create to capture the ID
		mockStore.On("Create", ctx, mock.Anything).Run(func(args mock.Arguments) {
			m := args.Get(1).(*model.Meeting)
			generatedID = m.Id
		}).Return(nil)

		token, _, err := svc.CreateMeeting(ctx, 1, "Setup", 3600, "http://base", nil)
		require.NoError(t, err)

		// Remove the call expectation so it doesn't interfere (or we just use fresh mocks after this helper?
		// No, we need the svc to have the SAME encrypter/key so the token is valid for THAT svc.
		// But we want to reset expected calls on the mockStore.)
		mockStore.ExpectedCalls = nil
		mockStore.Calls = nil // Clear history too if needed, though mostly expected calls matter

		return token, generatedID
	}

	t.Run("Valid Token and Found", func(t *testing.T) {
		svc, mockStore := setupMeetingService(t)
		ctx := context.Background()
		token, generatedID := getValidTokenAndID(t, svc, mockStore)

		validMeeting := &model.Meeting{
			Id:        generatedID,
			DomainId:  1,
			ExpiresAt: time.Now().Unix() + 3600,
		}

		mockStore.On("Get", ctx, generatedID).Return(validMeeting, nil)

		meeting, err := svc.GetMeeting(ctx, token)
		require.NoError(t, err)
		assert.NotNil(t, meeting)
		assert.Equal(t, generatedID, meeting.Id)
		mockStore.AssertExpectations(t)
	})

	t.Run("Valid Token but Not Found", func(t *testing.T) {
		svc, mockStore := setupMeetingService(t)
		ctx := context.Background()
		token, generatedID := getValidTokenAndID(t, svc, mockStore)

		mockStore.On("Get", ctx, generatedID).Return(nil, nil)

		meeting, err := svc.GetMeeting(ctx, token)
		require.NoError(t, err)
		assert.Nil(t, meeting)
		mockStore.AssertExpectations(t)
	})

	t.Run("Expired Meeting", func(t *testing.T) {
		svc, mockStore := setupMeetingService(t)
		ctx := context.Background()
		token, generatedID := getValidTokenAndID(t, svc, mockStore)

		expiredMeeting := &model.Meeting{
			Id:        generatedID,
			DomainId:  1,
			ExpiresAt: time.Now().Unix() - 100, // Expired
		}
		mockStore.On("Get", ctx, generatedID).Return(expiredMeeting, nil)

		meeting, err := svc.GetMeeting(ctx, token)
		require.Error(t, err)
		assert.Nil(t, meeting)
		assert.Contains(t, err.Error(), "expired")
		mockStore.AssertExpectations(t)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		svc, _ := setupMeetingService(t)
		ctx := context.Background()
		meeting, err := svc.GetMeeting(ctx, "invalid-token-string")
		require.Error(t, err)
		assert.Nil(t, meeting)
	})
}
