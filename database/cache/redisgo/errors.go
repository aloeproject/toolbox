package redisgo

import (
	"errors"
)

var (
	ErrConnExhausted = errors.New("redis: connection exhausted, please retry")
	ErrTimeout       = errors.New("redis: i/o timeout, please retry")
	ErrKeyNoExist    = errors.New("key does not exist")
)
