package app

import (
	"context"
	"encoding/json"

	"sherd-proof/internal/domain"
	"sherd-proof/internal/store"
)

func (s *Service) Review(ctx context.Context, command ReviewCommand) (*CaseView, error) {
	return s.updateCase(ctx, command.CaseID, command.RequestID, command, command.ExpectedRevision, command.ReviewerID, "case.reviewed",
		func(c *domain.ReconstructionCase, now TimeAlias) error {
			return c.Review(command.ReviewerID, command.Decision, command.Reason, command.ReopenKeys, now.Time)
		})
}

func (s *Service) Finalize(ctx context.Context, command FinalizeCommand) (*DossierView, error) {
	if err := requireRequestID(command.RequestID); err != nil {
		return nil, err
	}
	unlock := s.locks.lock(command.CaseID)
	defer unlock()
	data, err := commandBytes(command)
	if err != nil {
		return nil, err
	}
	replayed, err := s.store.Replay(ctx, command.RequestID, data)
	if err != nil {
		return nil, mapStoreError(err)
	}
	if replayed != nil {
		var view DossierView
		if err := json.Unmarshal(replayed.Body, &view); err != nil {
			return nil, err
		}
		return &view, nil
	}
	c, err := s.store.Load(ctx, command.CaseID)
	if err != nil {
		return nil, mapStoreError(err)
	}
	if err := requireRevision(c.Revision, command.ExpectedRevision); err != nil {
		return nil, err
	}
	now := s.now()
	dossier, err := s.store.Finalize(ctx, c, command.ExpectedRevision, command.DossierID,
		store.EventInput{Kind: "case.sealed", ActorID: command.ActorID, OccurredAt: now, Payload: map[string]string{"dossier_id": command.DossierID}}, now)
	if err != nil {
		return nil, mapStoreError(err)
	}
	view := dossierView(dossier, true, "档案载荷与事件摘要链校验通过")
	result, err := json.Marshal(view)
	if err == nil {
		err = s.store.Remember(ctx, command.RequestID, data, 200, result)
	}
	return &view, mapStoreError(err)
}
