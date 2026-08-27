package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"sherd-proof/internal/domain"
	"sherd-proof/internal/store"
)

type Service struct {
	store *store.SQLiteStore
	locks *keyedLocks
	now   func() time.Time
}

func NewService(repository *store.SQLiteStore) *Service {
	return &Service{store: repository, locks: newKeyedLocks(), now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Store() *store.SQLiteStore { return s.store }

func commandBytes(command any) ([]byte, error) { return json.Marshal(command) }

func requireRequestID(requestID string) error {
	if strings.TrimSpace(requestID) == "" {
		return domain.NewError(domain.CodeValidation, "request_id", "写操作必须提供 request_id")
	}
	return nil
}

func (s *Service) replayCase(ctx context.Context, requestID string, command any) (*CaseView, []byte, error) {
	data, err := commandBytes(command)
	if err != nil {
		return nil, nil, err
	}
	replayed, err := s.store.Replay(ctx, requestID, data)
	if err != nil {
		return nil, data, mapStoreError(err)
	}
	if replayed == nil {
		return nil, data, nil
	}
	var view CaseView
	if err := json.Unmarshal(replayed.Body, &view); err != nil {
		return nil, data, err
	}
	return &view, data, nil
}

func (s *Service) rememberCase(ctx context.Context, requestID string, command, view any) error {
	data, err := commandBytes(command)
	if err != nil {
		return err
	}
	result, err := json.Marshal(view)
	if err != nil {
		return err
	}
	return mapStoreError(s.store.Remember(ctx, requestID, data, 200, result))
}

func mapStoreError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, store.ErrNotFound):
		return domain.NewError(domain.CodeNotFound, "case_id", "案件不存在")
	case errors.Is(err, store.ErrRevisionConflict):
		return domain.NewError(domain.CodeConflict, "expected_revision", "页面修订号已过期，请刷新后重试")
	case errors.Is(err, store.ErrIdempotencyConflict):
		return domain.NewError(domain.CodeConflict, "request_id", "request_id 已用于不同请求载荷")
	default:
		return err
	}
}

func requireRevision(actual, expected int64) error {
	if expected <= 0 {
		return domain.NewError(domain.CodeValidation, "expected_revision", "必须提供正数修订号")
	}
	if actual != expected {
		return domain.NewError(domain.CodeConflict, "expected_revision", "页面修订号已过期，请刷新后重试")
	}
	return nil
}
