package app

import (
	"context"
	"encoding/json"

	"sherd-proof/internal/domain"
)

func (s *Service) RaiseChallenge(ctx context.Context, command RaiseChallengeCommand) (*CaseView, error) {
	return s.updateCase(ctx, command.CaseID, command.RequestID, command, command.ExpectedRevision, command.ActorID, "challenge.raised",
		func(c *domain.ReconstructionCase, now TimeAlias) error {
			return c.RaiseChallenge(domain.Challenge{ChallengeID: command.ChallengeID, HypothesisID: command.HypothesisID,
				EvidenceKey: command.EvidenceKey, RaisedBy: command.ActorID, Statement: command.Statement}, now.Time)
		})
}

func decodeReplacement(key string, raw json.RawMessage) (any, error) {
	if len(raw) == 0 {
		return nil, domain.NewError(domain.CodeValidation, "replacement", "补证动作必须提供替换内容")
	}
	switch key {
	case "edge_match", "fabric_match", "decoration_continuity":
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, domain.NewError(domain.CodeValidation, "replacement", "替换内容类型不正确")
		}
		return value, nil
	case "scale_measurements":
		var value map[string]float64
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, domain.NewError(domain.CodeValidation, "replacement", "测量值格式不正确")
		}
		return value, nil
	case "image_refs":
		var value []string
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, domain.NewError(domain.CodeValidation, "replacement", "图像引用格式不正确")
		}
		return value, nil
	default:
		return nil, domain.NewError(domain.CodeValidation, "evidence_key", "未知证据项")
	}
}

func (s *Service) ResolveChallenge(ctx context.Context, command ResolveChallengeCommand) (*CaseView, error) {
	return s.updateCase(ctx, command.CaseID, command.RequestID, command, command.ExpectedRevision, command.ActorID, "challenge.resolved",
		func(c *domain.ReconstructionCase, now TimeAlias) error {
			challenge := c.Challenges[command.ChallengeID]
			var replacement any
			var err error
			if command.Kind == domain.ResolutionSupplement {
				if challenge == nil {
					return domain.NewError(domain.CodeNotFound, "challenge_id", "异议不存在")
				}
				replacement, err = decodeReplacement(challenge.EvidenceKey, command.Replacement)
				if err != nil {
					return err
				}
			}
			return c.ResolveChallenge(command.ChallengeID, command.Kind, command.Note, command.ActorID, replacement, now.Time)
		})
}

func (s *Service) AdvanceToReview(ctx context.Context, command CaseCommand) (*CaseView, error) {
	return s.updateCase(ctx, command.CaseID, command.RequestID, command, command.ExpectedRevision, command.ActorID, "case.review_requested",
		func(c *domain.ReconstructionCase, now TimeAlias) error { return c.AdvanceToReview(now.Time) })
}
