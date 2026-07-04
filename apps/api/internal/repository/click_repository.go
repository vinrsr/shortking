package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"shortking-api/internal/models"
)

type ClickRepository interface {
	CreateBatch(ctx context.Context, events []models.ClickEvent) error
	ListByLink(ctx context.Context, linkID uuid.UUID, limit int) ([]models.ClickEvent, error)
	CountAll(ctx context.Context) (int64, error)
}

type clickRepository struct {
	db *gorm.DB
}

func NewClickRepository(db *gorm.DB) ClickRepository {
	return &clickRepository{db: db}
}

func (r *clickRepository) CreateBatch(ctx context.Context, events []models.ClickEvent) error {
	if len(events) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(events, 100).Error
}

func (r *clickRepository) ListByLink(ctx context.Context, linkID uuid.UUID, limit int) ([]models.ClickEvent, error) {
	var events []models.ClickEvent
	err := r.db.WithContext(ctx).
		Where("link_id = ?", linkID).
		Order("clicked_at DESC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *clickRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.ClickEvent{}).Count(&count).Error
	return count, err
}
