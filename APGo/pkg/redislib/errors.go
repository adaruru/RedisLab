package redislib

import "errors"

var (
	// ErrInvalidRedisMode 無效的 Redis 模式
	ErrInvalidRedisMode = errors.New("invalid redis mode")
	// ErrKeyNotFound Key 不存在
	ErrKeyNotFound = errors.New("key not found")
	// ErrConnectionFailed 連線失敗
	ErrConnectionFailed = errors.New("connection failed")
	// ErrWriteFailed 寫入失敗
	ErrWriteFailed = errors.New("write failed")
	// ErrReadFailed 讀取失敗
	ErrReadFailed = errors.New("read failed")
)
