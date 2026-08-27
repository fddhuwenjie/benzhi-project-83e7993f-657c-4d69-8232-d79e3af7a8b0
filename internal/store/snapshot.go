package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"sherd-proof/internal/domain"
)

func writeAggregate(ctx context.Context, tx *sql.Tx, c *domain.ReconstructionCase, creating bool) error {
	state, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if creating {
		_, err = tx.ExecContext(ctx, `INSERT INTO cases(case_id,status,revision,state_json,created_at,updated_at) VALUES(?,?,?,?,?,?)`,
			c.CaseID, c.Status, c.Revision, state, c.CreatedAt.UTC().Format(timeFormat), c.UpdatedAt.UTC().Format(timeFormat))
	} else {
		result, execErr := tx.ExecContext(ctx, `UPDATE cases SET status=?, revision=?, state_json=?, updated_at=? WHERE case_id=?`,
			c.Status, c.Revision, state, c.UpdatedAt.UTC().Format(timeFormat), c.CaseID)
		if execErr != nil {
			return execErr
		}
		count, countErr := result.RowsAffected()
		if countErr != nil {
			return countErr
		}
		if count != 1 {
			return ErrNotFound
		}
	}
	if err != nil {
		return err
	}
	return rewriteChildren(ctx, tx, c)
}

func rewriteChildren(ctx context.Context, tx *sql.Tx, c *domain.ReconstructionCase) error {
	for _, table := range []string{"evidence_versions", "challenges", "reviews", "hypotheses", "sherds"} {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE case_id = ?", table), c.CaseID); err != nil {
			return err
		}
	}
	for _, sherd := range c.Sherds {
		data, err := marshal(sherd)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO sherds(case_id,sherd_id,record_json) VALUES(?,?,?)`, c.CaseID, sherd.SherdID, data); err != nil {
			return err
		}
	}
	for _, h := range c.Hypotheses {
		data, err := marshal(h)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO hypotheses(case_id,hypothesis_id,status,author_id,record_json) VALUES(?,?,?,?,?)`, c.CaseID, h.HypothesisID, h.Status, h.AuthorID, data); err != nil {
			return err
		}
		for _, version := range h.EvidenceVersions {
			evidence, err := marshal(version.Evidence)
			if err != nil {
				return err
			}
			if _, err = tx.ExecContext(ctx, `INSERT INTO evidence_versions(case_id,hypothesis_id,version,evidence_json,changed_by,note,created_at) VALUES(?,?,?,?,?,?,?)`,
				c.CaseID, h.HypothesisID, version.Version, evidence, version.ChangedBy, version.Note, version.CreatedAt.UTC().Format(timeFormat)); err != nil {
				return err
			}
		}
	}
	for _, challenge := range c.Challenges {
		data, err := marshal(challenge)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO challenges(case_id,challenge_id,hypothesis_id,status,record_json) VALUES(?,?,?,?,?)`,
			c.CaseID, challenge.ChallengeID, challenge.HypothesisID, challenge.Status, data); err != nil {
			return err
		}
	}
	for index, review := range c.Reviews {
		data, err := marshal(review)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO reviews(case_id,review_index,record_json) VALUES(?,?,?)`, c.CaseID, index, data); err != nil {
			return err
		}
	}
	return nil
}
