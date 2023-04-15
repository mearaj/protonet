package model

import (
	"fmt"
	"time"
)

type Contact struct {
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Avatar           []byte
	Identified       bool
	PublicKey        string
	AccountPublicKey string
}

func (c *Contact) GetDBFullKey() (key string, err error) {
	if len(c.PublicKey) == 0 || len(c.AccountPublicKey) == 0 || c.UpdatedAt.IsZero() || c.CreatedAt.IsZero() {
		return key, ErrInvalidContact
	}
	updatedTime := c.UpdatedAt.UnixNano()
	createdTime := c.CreatedAt.UnixNano()
	key = fmt.Sprintf("%s%s%s%s%s%s%d%s%d",
		KeyPrefixContacts,
		KeySeparator, c.AccountPublicKey,
		KeySeparator, c.PublicKey,
		KeySeparator, updatedTime,
		KeySeparator, createdTime,
	)
	return key, nil
}

func (c *Contact) GetDBPrefixKey() (key string) {
	key = KeyPrefixContacts
	if len(c.AccountPublicKey) == 0 {
		return key
	}
	key = fmt.Sprintf("%s%s%s", key, KeySeparator, c.AccountPublicKey)
	if len(c.PublicKey) == 0 {
		return key
	}
	key = fmt.Sprintf("%s%s%s", key, KeySeparator, c.PublicKey)
	if c.UpdatedAt.IsZero() {
		return key
	}
	key = fmt.Sprintf("%s%s%d", key, KeySeparator, c.UpdatedAt.UnixNano())
	if c.CreatedAt.IsZero() {
		return key
	}
	key = fmt.Sprintf("%s%s%d", key, KeySeparator, c.CreatedAt.UnixNano())
	return key
}
