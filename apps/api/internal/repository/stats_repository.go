package repository

import (
	"context"

	"gorm.io/gorm"
)

type statsRow struct {
	ID            int   `gorm:"primaryKey"`
	QRGenerations int64 `gorm:"column:qr_generations"`
}

func (statsRow) TableName() string { return "stats" }

type StatsRepository interface {
	IncrementQRGenerations(ctx context.Context) error
	TotalQRGenerations(ctx context.Context) (int64, error)
}

type statsRepository struct {
	db *gorm.DB
}

func NewStatsRepository(db *gorm.DB) StatsRepository {
	return &statsRepository{db: db}
}

func (r *statsRepository) IncrementQRGenerations(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Model(&statsRow{}).
		Where("id = 1").
		UpdateColumn("qr_generations", gorm.Expr("qr_generations + 1")).Error
}

func (r *statsRepository) TotalQRGenerations(ctx context.Context) (int64, error) {
	var row statsRow
	err := r.db.WithContext(ctx).First(&row, "id = 1").Error
	if err != nil {
		return 0, err
	}
	return row.QRGenerations, nil
}
