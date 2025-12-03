package model

import (
	"encoding/json"
	"fmt"
)

type CallHangupData struct {
	Cause     *string `json:"cause"`
	MeetingId *string `json:"meeting_id,omitempty"`
}

type Call struct {
	Id    string `json:"id"`
	AppId string `json:"app_id"`

	Data    CallHangupData  `json:"-"`
	RawData json.RawMessage `json:"data"`
}

func (e *Call) UnmarshalJSON(data []byte) error {
	type Alias Call
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var dataStr string
	if err := json.Unmarshal(e.RawData, &dataStr); err == nil {
		if err := json.Unmarshal([]byte(dataStr), &e.Data); err != nil {
			return fmt.Errorf("failed to unmarshal nested data JSON: %w", err)
		}
	} else {
		if err := json.Unmarshal(e.RawData, &e.Data); err != nil {
			return fmt.Errorf("failed to unmarshal data JSON object: %w", err)
		}
	}

	return nil
}

func CallFromJson(js []byte) (*Call, error) {
	var call *Call
	err := json.Unmarshal(js, &call)
	return call, err

}
