package app

import (
	"context"
	"time"

	"sherd-proof/internal/domain"
	"sherd-proof/internal/store"
)

func (s *Service) CreateCase(ctx context.Context, command CreateCaseCommand) (*CaseView, error) {
	if err := requireRequestID(command.RequestID); err != nil {
		return nil, err
	}
	if command.ExpectedRevision != 0 {
		return nil, domain.NewError(domain.CodeValidation, "expected_revision", "创建案件的 expected_revision 必须为 0")
	}
	unlock := s.locks.lock(command.CaseID)
	defer unlock()
	if replay, _, err := s.replayCase(ctx, command.RequestID, command); err != nil || replay != nil {
		return replay, err
	}
	now := s.now()
	c, err := domain.NewCase(command.CaseID, command.SiteUnit, command.VesselClass, command.OwnerID, now)
	if err != nil {
		return nil, err
	}
	err = s.store.Create(ctx, c, store.EventInput{Kind: "case.created", ActorID: command.OwnerID, OccurredAt: now,
		Payload: map[string]any{"site_unit": c.SiteUnit, "vessel_class": c.VesselClass}})
	if err != nil {
		return nil, mapStoreError(err)
	}
	view, err := s.caseView(ctx, c)
	if err == nil {
		err = s.rememberCase(ctx, command.RequestID, command, view)
	}
	return view, err
}

func (s *Service) AddSherd(ctx context.Context, command AddSherdCommand) (*CaseView, error) {
	return s.updateCase(ctx, command.CaseID, command.RequestID, command, command.ExpectedRevision, command.ActorID, "sherd.added",
		func(c *domain.ReconstructionCase, nowTime TimeAlias) error {
			return c.AddSherd(command.Sherd, command.ActorID, nowTime.Time)
		})
}

func (s *Service) FreezeBaseline(ctx context.Context, command CaseCommand) (*CaseView, error) {
	return s.updateCase(ctx, command.CaseID, command.RequestID, command, command.ExpectedRevision, command.ActorID, "baseline.frozen",
		func(c *domain.ReconstructionCase, now TimeAlias) error {
			return c.FreezeBaseline(command.ActorID, now.Time)
		})
}

type TimeAlias struct{ Time time.Time }

type caseMutation func(*domain.ReconstructionCase, TimeAlias) error

func (s *Service) updateCase(ctx context.Context, caseID, requestID string, command any, expected int64, actor, eventKind string, mutate caseMutation) (*CaseView, error) {
	if err := requireRequestID(requestID); err != nil {
		return nil, err
	}
	unlock := s.locks.lock(caseID)
	defer unlock()
	if replay, _, err := s.replayCase(ctx, requestID, command); err != nil || replay != nil {
		return replay, err
	}
	c, err := s.store.Load(ctx, caseID)
	if err != nil {
		return nil, mapStoreError(err)
	}
	if err := requireRevision(c.Revision, expected); err != nil {
		return nil, err
	}
	now := s.now()
	if err := mutate(c, TimeAlias{Time: now}); err != nil {
		return nil, err
	}
	err = s.store.Save(ctx, c, expected, store.EventInput{Kind: eventKind, ActorID: actor, OccurredAt: now, Payload: command})
	if err != nil {
		return nil, mapStoreError(err)
	}
	view, err := s.caseView(ctx, c)
	if err == nil {
		err = s.rememberCase(ctx, requestID, command, view)
	}
	return view, err
}
