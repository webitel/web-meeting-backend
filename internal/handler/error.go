package handler

import (
	"encoding/json"
	"net/http"
)

type HttpError struct {
	Id     string `json:"id"`
	Code   uint32 `json:"code"`
	Detail string `json:"detail"`
}

func (e *HttpError) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func NewHttpError(code uint32, id, detail string) *HttpError {
	return &HttpError{
		Id:     id,
		Code:   code,
		Detail: detail,
	}
}

func NewBadRequest(id string, err error) *HttpError {
	return NewHttpError(http.StatusBadRequest, id, err.Error())
}
