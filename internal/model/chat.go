package model

type ChatCloseInfo struct {
	ConversationId string `json:"conversation_id" db:"conversation_id"`
	CloserId       string `json:"closer_id" db:"closer_id"`
	AuthUserId     int64  `json:"auth_user_id" db:"auth_user_id"`
}
