package services

import (
	"context"
	"time"

	"manticore-imports/internal/domain"
)

type MetricsService struct {
	requests requestRepository
}

func NewMetricsService(requests requestRepository) *MetricsService {
	return &MetricsService{requests: requests}
}

func (s *MetricsService) Snapshot(ctx context.Context) (*domain.MetricsSnapshot, error) {
	byStatus, err := s.requests.CountByStatus(ctx)
	if err != nil {
		return nil, err
	}
	total := 0
	for _, count := range byStatus {
		total += count
	}
	return &domain.MetricsSnapshot{
		TotalRequests: total,
		ByStatus:      byStatus,
		GeneratedAt:   time.Now().UTC(),
	}, nil
}
