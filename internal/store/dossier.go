package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"sherd-proof/internal/domain"
)

func (s *SQLiteStore) Finalize(ctx context.Context, c *domain.ReconstructionCase, expected int64, dossierID string, event EventInput, sealedAt time.Time) (*domain.FinalDossier, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	current, err := currentRevision(ctx, tx, c.CaseID)
	if err != nil {
		return nil, err
	}
	if current != expected {
		return nil, ErrRevisionConflict
	}
	head, err := appendEvent(ctx, tx, c.CaseID, event)
	if err != nil {
		return nil, err
	}
	dossier, err := domain.BuildDossier(c, dossierID, head, sealedAt)
	if err != nil {
		return nil, err
	}
	if err := c.Seal(sealedAt); err != nil {
		return nil, err
	}
	c.Revision = expected + 1
	if err := writeAggregate(ctx, tx, c, false); err != nil {
		return nil, err
	}
	data, err := json.Marshal(dossier)
	if err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO dossiers(case_id,dossier_id,dossier_json,sha256,event_chain_head,sealed_at) VALUES(?,?,?,?,?,?)`,
		c.CaseID, dossier.DossierID, data, dossier.SHA256, dossier.EventChainHead, sealedAt.UTC().Format(timeFormat))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	s.cacheCase(c)
	return dossier, nil
}

func (s *SQLiteStore) Dossier(ctx context.Context, caseID string) (*domain.FinalDossier, error) {
	var data []byte
	err := s.db.QueryRowContext(ctx, `SELECT dossier_json FROM dossiers WHERE case_id=?`, caseID).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var dossier domain.FinalDossier
	if err := json.Unmarshal(data, &dossier); err != nil {
		return nil, err
	}
	return &dossier, nil
}

func (s *SQLiteStore) VerifyDossier(ctx context.Context, caseID string) (bool, string, error) {
	dossier, err := s.Dossier(ctx, caseID)
	if err != nil {
		return false, "", err
	}
	head, err := s.VerifyEventChain(ctx, caseID)
	if err != nil {
		return false, "事件摘要链校验失败", nil
	}
	valid, message := domain.VerifyDossier(dossier, head)
	return valid, message, nil
}
