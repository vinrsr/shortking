package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email           string    `gorm:"uniqueIndex;not null"`
	PasswordHash    string    `gorm:"not null"`
	DisplayName     string    `gorm:"not null;default:''"`
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Links           []Link `gorm:"foreignKey:UserID"`
}

func (User) TableName() string { return "users" }

type Link struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        *uuid.UUID `gorm:"type:uuid;index"` // nil for anonymous (no-login) links
	ShortCode     string     `gorm:"uniqueIndex;size:32;not null"`
	Destination   string     `gorm:"not null"`
	ExpiresAt     *time.Time
	MaxClicks     *int
	ClickCount    int  `gorm:"not null;default:0"`
	IsActive      bool `gorm:"not null;default:true"`
	QRGeneratedAt *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	User          User         `gorm:"foreignKey:UserID"`
	Clicks        []ClickEvent `gorm:"foreignKey:LinkID"`
}

func (Link) TableName() string { return "links" }

type ClickEvent struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	LinkID    uuid.UUID `gorm:"type:uuid;index;not null" json:"linkId"`
	ClickedAt time.Time `gorm:"index;not null" json:"clickedAt"`
	Referrer  string    `gorm:"not null;default:''" json:"referrer"`
	UserAgent string    `gorm:"not null;default:''" json:"userAgent"`
	IPHash    string    `gorm:"not null;default:''" json:"-"`
}

func (ClickEvent) TableName() string { return "click_events" }
