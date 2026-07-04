package service

import (
	"context"

	"shortking-api/internal/repository"
)

type StatsService struct {
	stats repository.StatsRepository
}

func NewStatsService(stats repository.StatsRepository) *StatsService {
	return &StatsService{stats: stats}
}

func (s *StatsService) RecordQRGeneration(ctx context.Context) error {
	return s.stats.IncrementQRGenerations(ctx)
}

func (s *StatsService) TotalQRGenerations(ctx context.Context) (int64, error) {
	return s.stats.TotalQRGenerations(ctx)
}
