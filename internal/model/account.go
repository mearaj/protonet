package model

import (
	"fmt"
	"time"
)

type Account struct {
	CreatedAt    time.Time
	UpdatedAt    time.Time
	PrivateKey   string
	PublicImage  []byte
	PrivateImage []byte
	PublicKey    string
	EthAddress   string
}

func (a *Account) GetDBFullKey() (key string, err error) {
	if len(a.PublicKey) == 0 || a.UpdatedAt.IsZero() || a.CreatedAt.IsZero() {
		return key, ErrInvalidAccount
	}
	updatedAt := a.UpdatedAt.UnixNano()
	createdAt := a.CreatedAt.UnixNano()
	key = fmt.Sprintf("%s%s%d%s%d%s%s",
		KeyPrefixAccounts,
		KeySeparator, updatedAt,
		KeySeparator, createdAt,
		KeySeparator, a.PublicKey,
	)
	return key, nil
}

func (a *Account) GetAccountDBPrefixKey() (key string) {
	key = KeyPrefixAccounts
	if a.UpdatedAt.IsZero() {
		return key
	}
	updatedAt := a.UpdatedAt.UnixMilli()
	key = fmt.Sprintf("%s%s%d", key, KeySeparator, updatedAt)
	if a.CreatedAt.IsZero() {
		return key
	}
	createdAt := a.CreatedAt.UnixMilli()
	key = fmt.Sprintf("%s%s%d", key, KeySeparator, createdAt)
	if len(a.PublicKey) == 0 {
		return key
	}
	key = fmt.Sprintf("%s%s%s", key, KeySeparator, a.PublicKey)
	if a.UpdatedAt.IsZero() {
		return key
	}
	return key
}
