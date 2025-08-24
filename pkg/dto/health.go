package dto

import "time"

type HealthStatus struct {
	Status    string        `json:"status"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration_ms"`
}
