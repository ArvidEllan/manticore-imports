package domain

import "time"

type MetricsSnapshot struct {
	TotalRequests int            `json:"totalRequests"`
	ByStatus      map[string]int `json:"byStatus"`
	GeneratedAt   time.Time      `json:"generatedAt"`
}
