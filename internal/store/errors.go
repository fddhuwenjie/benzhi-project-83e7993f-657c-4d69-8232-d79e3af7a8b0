package store

import "errors"

var (
	ErrNotFound            = errors.New("记录不存在")
	ErrRevisionConflict    = errors.New("修订号冲突")
	ErrIdempotencyConflict = errors.New("request_id 已用于不同载荷")
)
