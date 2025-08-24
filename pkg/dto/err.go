package dto

import (
	"encoding/json"
	"time"
)

type ErrorResponse struct {
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

func NewErr(msg string) ErrorResponse {
	return ErrorResponse{
		Message: msg,
		Time:    time.Now(),
	}
}

func (e ErrorResponse) ToString() string {
	b, err := json.MarshalIndent(e, "", "    ")
	if err != nil {
		return ""
	}

	return string(b)
}
