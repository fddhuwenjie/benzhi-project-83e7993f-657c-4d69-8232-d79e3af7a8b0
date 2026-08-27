package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

func (s *SQLiteStore) Replay(ctx context.Context, requestID string, requestBody []byte) (*IdempotentResult, error) {
	if requestID == "" {
		return nil, nil
	}
	var storedFingerprint string
	var result IdempotentResult
	err := s.db.QueryRowContext(ctx, `SELECT fingerprint,status_code,result_json FROM idempotency WHERE request_id=?`, requestID).
		Scan(&storedFingerprint, &result.StatusCode, &result.Body)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if storedFingerprint != fingerprint(requestBody) {
		return nil, ErrIdempotencyConflict
	}
	return &result, nil
}

func (s *SQLiteStore) Remember(ctx context.Context, requestID string, requestBody []byte, status int, result []byte) error {
	if requestID == "" {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO idempotency(request_id,fingerprint,status_code,result_json,created_at) VALUES(?,?,?,?,?)`,
		requestID, fingerprint(requestBody), status, result, time.Now().UTC().Format(timeFormat))
	if err != nil {
		replayed, replayErr := s.Replay(ctx, requestID, requestBody)
		if replayErr != nil {
			return replayErr
		}
		if replayed != nil {
			return nil
		}
	}
	return err
}
