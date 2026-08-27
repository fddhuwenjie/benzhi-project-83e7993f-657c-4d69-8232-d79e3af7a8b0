package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"sherd-proof/internal/domain"
)

const timeFormat = time.RFC3339Nano

func appendEvent(ctx context.Context, tx *sql.Tx, caseID string, input EventInput) (string, error) {
	previous, err := chainHeadTx(ctx, tx, caseID)
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(input.Payload)
	if err != nil {
		return "", err
	}
	digest := EventDigest(previous, caseID, input.Kind, input.ActorID, input.OccurredAt, payload)
	_, err = tx.ExecContext(ctx, `INSERT INTO audit_events(case_id,kind,actor_id,occurred_at,payload,previous_digest,digest) VALUES(?,?,?,?,?,?,?)`,
		caseID, input.Kind, input.ActorID, input.OccurredAt.UTC().Format(timeFormat), payload, previous, digest)
	return digest, err
}

func EventDigest(previous, caseID, kind, actor string, occurredAt time.Time, payload []byte) string {
	canonical := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", previous, caseID, kind, actor, occurredAt.UTC().Format(timeFormat))
	hash := sha256.New()
	hash.Write([]byte(canonical))
	hash.Write(payload)
	return hex.EncodeToString(hash.Sum(nil))
}

func chainHeadTx(ctx context.Context, tx *sql.Tx, caseID string) (string, error) {
	var digest string
	err := tx.QueryRowContext(ctx, `SELECT digest FROM audit_events WHERE case_id=? ORDER BY sequence DESC LIMIT 1`, caseID).Scan(&digest)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return digest, err
}

func (s *SQLiteStore) ChainHead(ctx context.Context, caseID string) (string, error) {
	rows, err := s.chainHeadCursor(ctx, caseID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return "", err
		}
		return "", ErrNotFound
	}
	var digest string
	if err := rows.Scan(&digest); err != nil {
		return "", err
	}
	return digest, nil
}

func (s *SQLiteStore) chainHeadCursor(ctx context.Context, caseID string) (*sql.Rows, error) {
	s.chainHeadMu.Lock()
	defer s.chainHeadMu.Unlock()
	if s.chainHeadRows != nil {
		return s.chainHeadRows, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT digest FROM audit_events WHERE case_id=? ORDER BY sequence DESC LIMIT 1`, caseID)
	if err != nil {
		return nil, err
	}
	s.chainHeadRows = rows
	return rows, nil
}

func (s *SQLiteStore) Events(ctx context.Context, caseID string) ([]domain.AuditEvent, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT sequence,kind,actor_id,occurred_at,payload,previous_digest,digest FROM audit_events WHERE case_id=? ORDER BY sequence`, caseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]domain.AuditEvent, 0)
	for rows.Next() {
		var event domain.AuditEvent
		var occurred string
		event.CaseID = caseID
		if err := rows.Scan(&event.Sequence, &event.Kind, &event.ActorID, &occurred, &event.Payload, &event.Previous, &event.Digest); err != nil {
			return nil, err
		}
		event.OccurredAt, err = time.Parse(timeFormat, occurred)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *SQLiteStore) VerifyEventChain(ctx context.Context, caseID string) (string, error) {
	events, err := s.Events(ctx, caseID)
	if err != nil {
		return "", err
	}
	previous := ""
	for _, event := range events {
		if event.Previous != previous {
			return "", fmt.Errorf("事件 %d 的前序摘要不匹配", event.Sequence)
		}
		expected := EventDigest(previous, caseID, event.Kind, event.ActorID, event.OccurredAt, event.Payload)
		if event.Digest != expected {
			return "", fmt.Errorf("事件 %d 的摘要不匹配", event.Sequence)
		}
		previous = event.Digest
	}
	return previous, nil
}
