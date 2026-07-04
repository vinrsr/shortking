package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"shortking-api/internal/models"
)

type LinkRepository interface {
	Create(ctx context.Context, link *models.Link) error
	FindByShortCode(ctx context.Context, code string) (*models.Link, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Link, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Link, error)
	Update(ctx context.Context, link *models.Link) error
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementClickCount(ctx context.Context, id uuid.UUID, by int) error
	CountAll(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)
}

type linkRepository struct {
	db *gorm.DB
}

func NewLinkRepository(db *gorm.DB) LinkRepository {
	return &linkRepository{db: db}
}

func (r *linkRepository) Create(ctx context.Context, link *models.Link) error {
	return translateError(r.db.WithContext(ctx).Create(link).Error)
}

func (r *linkRepository) FindByShortCode(ctx context.Context, code string) (*models.Link, error) {
	var link models.Link
	err := r.db.WithContext(ctx).Where("short_code = ?", code).First(&link).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *linkRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Link, error) {
	var link models.Link
	err := r.db.WithContext(ctx).First(&link, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *linkRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Link, error) {
	var links []models.Link
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&links).Error
	return links, err
}

func (r *linkRepository) Update(ctx context.Context, link *models.Link) error {
	return translateError(r.db.WithContext(ctx).Save(link).Error)
}

func (r *linkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Link{}, "id = ?", id).Error
}

func (r *linkRepository) IncrementClickCount(ctx context.Context, id uuid.UUID, by int) error {
	return r.db.WithContext(ctx).
		Model(&models.Link{}).
		Where("id = ?", id).
		UpdateColumn("click_count", gorm.Expr("click_count + ?", by)).Error
}

func (r *linkRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Link{}).Count(&count).Error
	return count, err
}

// CountActive counts links that are neither deactivated, date-expired, nor
// past their max-click limit, i.e. links that would still redirect if
// visited right now.
func (r *linkRepository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Link{}).
		Where("is_active = true").
		Where("expires_at IS NULL OR expires_at > now()").
		Where("max_clicks IS NULL OR click_count < max_clicks").
		Count(&count).Error
	return count, err
}
