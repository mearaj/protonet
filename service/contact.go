package service

import (
	"time"
)

type Contact struct {
	CreatedAt        time.Time
	UpdatedAt        time.Time
	PublicKey        string     `gorm:"primaryKey"`
	AccountPublicKey string     `gorm:"primaryKey;foreignKey"`
	Messages         []*Message `gorm:"foreignKey:AccountPublicKey,ContactPublicKey;References:AccountPublicKey,PublicKey;constraint:OnDelete:CASCADE"`
	Avatar           []byte
}
