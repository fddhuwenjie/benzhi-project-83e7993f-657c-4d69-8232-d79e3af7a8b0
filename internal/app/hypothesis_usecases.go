package app

import (
	"context"

	"sherd-proof/internal/domain"
)

func (s *Service) AddHypothesis(ctx context.Context, command AddHypothesisCommand) (*CaseView, error) {
	return s.updateCase(ctx, command.CaseID, command.RequestID, command, command.ExpectedRevision, command.ActorID, "hypothesis.added",
		func(c *domain.ReconstructionCase, now TimeAlias) error {
			return c.AddHypothesis(domain.JoinHypothesis{HypothesisID: command.HypothesisID, SherdIDs: command.SherdIDs, Evidence: command.Evidence}, command.ActorID, now.Time)
		})
}

func (s *Service) ReviseReturnedEvidence(ctx context.Context, command ReviseEvidenceCommand) (*CaseView, error) {
	return s.updateCase(ctx, command.CaseID, command.RequestID, command, command.ExpectedRevision, command.ActorID, "hypothesis.returned_evidence_revised",
		func(c *domain.ReconstructionCase, now TimeAlias) error {
			replacement, err := decodeReplacement(command.EvidenceKey, command.Replacement)
			if err != nil {
				return err
			}
			return c.ReviseReturnedEvidence(command.HypothesisID, command.EvidenceKey, replacement, command.ActorID, command.Note, now.Time)
		})
}

func (s *Service) SubmitHypothesis(ctx context.Context, command HypothesisCommand) (*CaseView, error) {
	return s.updateCase(ctx, command.CaseID, command.RequestID, command, command.ExpectedRevision, command.ActorID, "hypothesis.published",
		func(c *domain.ReconstructionCase, now TimeAlias) error {
			return c.SubmitHypothesis(command.HypothesisID, command.ActorID, now.Time)
		})
}
