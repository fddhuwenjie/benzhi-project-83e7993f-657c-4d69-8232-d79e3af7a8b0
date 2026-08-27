package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"sherd-proof/internal/domain"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct{ db *sql.DB }

type EventInput struct {
	Kind       string
	ActorID    string
	OccurredAt time.Time
	Payload    any
}

type IdempotentResult struct {
	StatusCode int
	Body       []byte
}

func Open(path string) (*SQLiteStore, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("数据库路径不能为空")
	}
	dsn := path
	if path != ":memory:" && !strings.HasPrefix(path, "file:") {
		dsn = "file:" + path
	}
	separator := "?"
	if strings.Contains(dsn, "?") {
		separator = "&"
	}
	db, err := sql.Open("sqlite", dsn+separator+"_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(4)
	if _, err = db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("初始化数据库: %w", err)
	}
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Close() error { return s.db.Close() }

func (s *SQLiteStore) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }

func (s *SQLiteStore) Create(ctx context.Context, c *domain.ReconstructionCase, event EventInput) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	c.Revision = 1
	if err := writeAggregate(ctx, tx, c, true); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return domain.NewError(domain.CodeConflict, "case_id", "案件编号已存在")
		}
		return err
	}
	if _, err := appendEvent(ctx, tx, c.CaseID, event); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SQLiteStore) Load(ctx context.Context, caseID string) (*domain.ReconstructionCase, error) {
	var data []byte
	err := s.db.QueryRowContext(ctx, `SELECT state_json FROM cases WHERE case_id = ?`, caseID).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var c domain.ReconstructionCase
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("解析案件快照: %w", err)
	}
	normalizeMaps(&c)
	return &c, nil
}

func (s *SQLiteStore) Save(ctx context.Context, c *domain.ReconstructionCase, expected int64, event EventInput) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	current, err := currentRevision(ctx, tx, c.CaseID)
	if err != nil {
		return err
	}
	if current != expected {
		return ErrRevisionConflict
	}
	c.Revision = expected + 1
	if err := writeAggregate(ctx, tx, c, false); err != nil {
		return err
	}
	if _, err := appendEvent(ctx, tx, c.CaseID, event); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SQLiteStore) List(ctx context.Context, limit int) ([]*domain.ReconstructionCase, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `SELECT state_json FROM cases ORDER BY updated_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*domain.ReconstructionCase, 0)
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var c domain.ReconstructionCase
		if err := json.Unmarshal(data, &c); err != nil {
			return nil, err
		}
		normalizeMaps(&c)
		result = append(result, &c)
	}
	return result, rows.Err()
}

func currentRevision(ctx context.Context, tx *sql.Tx, caseID string) (int64, error) {
	var revision int64
	err := tx.QueryRowContext(ctx, `SELECT revision FROM cases WHERE case_id = ?`, caseID).Scan(&revision)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return revision, err
}

func normalizeMaps(c *domain.ReconstructionCase) {
	if c.Sherds == nil {
		c.Sherds = map[string]*domain.SherdRecord{}
	}
	if c.Hypotheses == nil {
		c.Hypotheses = map[string]*domain.JoinHypothesis{}
	}
	if c.Challenges == nil {
		c.Challenges = map[string]*domain.Challenge{}
	}
	if c.Editors == nil {
		c.Editors = map[string]bool{}
	}
	if c.ReopenedKeys == nil {
		c.ReopenedKeys = map[string]bool{}
	}
	for _, h := range c.Hypotheses {
		if h.LockedKeys == nil {
			h.LockedKeys = map[string]bool{}
		}
	}
}

func marshal(value any) ([]byte, error) { return json.Marshal(value) }

func fingerprint(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
