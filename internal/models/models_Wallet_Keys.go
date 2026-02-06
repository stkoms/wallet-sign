package models

import (
	"time"
)

type WalletKey struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Address      string    `gorm:"size:128;uniqueIndex" json:"address"`
	KeyType      string    `gorm:"size:32" json:"keyType"`
	EncryptedKey []byte    `gorm:"type:blob" json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (WalletKey) TableName() string { return "wallet_keys" }
