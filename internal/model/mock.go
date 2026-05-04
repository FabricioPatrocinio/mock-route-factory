package model

import (
	"encoding/json"
	"time"
)

type Mock struct {
	ID           int64           `json:"id"`
	Method       string          `json:"method"`
	Path         string          `json:"path"`
	Status       int             `json:"status"`
	ResponseBody json.RawMessage `json:"response"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
