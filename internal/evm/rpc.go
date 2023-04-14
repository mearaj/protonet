package evm

import (
	"errors"
	"strings"
)

type RPC string

var (
	ErrAPIKeyIsRequired = errors.New("api key is required")
	ErrInvalidRPC       = errors.New("invalid rpc")
)

func (r RPC) KeyRequired() bool {
	return strings.Contains(strings.ToLower(string(r)), "api_key")
}

func (r RPC) GetURL(key string) (string, error) {
	if !r.KeyRequired() {
		return string(r), nil
	}
	if r.KeyRequired() && len(key) == 0 {
		return string(r), ErrAPIKeyIsRequired
	}
	lastIndex := strings.LastIndex(string(r), "/")
	if lastIndex <= 0 {
		return string(r), ErrInvalidRPC
	}
	return string(r)[:lastIndex], nil
}
