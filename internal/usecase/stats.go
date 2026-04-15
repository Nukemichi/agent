package usecase

import (
	"context"
	"fmt"

	"agent-michi/internal/domain"
)

// StatsUseCase delegates stats collection to a StatsCollector.
type StatsUseCase struct {
	collector domain.StatsCollector
}

// NewStatsUseCase creates a new StatsUseCase.
func NewStatsUseCase(collector domain.StatsCollector) *StatsUseCase {
	return &StatsUseCase{collector: collector}
}

// GetStats returns current system statistics.
func (s *StatsUseCase) GetStats(ctx context.Context) (domain.StatsResponse, error) {
	stats, err := s.collector.Collect(ctx)
	if err != nil {
		return domain.StatsResponse{}, fmt.Errorf("collect stats: %w", err)
	}
	return stats, nil
}
