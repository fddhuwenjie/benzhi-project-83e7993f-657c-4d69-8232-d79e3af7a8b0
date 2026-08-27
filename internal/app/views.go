package app

import (
	"context"
	"sort"

	"sherd-proof/internal/domain"
)

type ActionSet struct {
	CanAddSherd       bool `json:"can_add_sherd"`
	CanFreezeBaseline bool `json:"can_freeze_baseline"`
	CanAddHypothesis  bool `json:"can_add_hypothesis"`
	CanChallenge      bool `json:"can_challenge"`
	CanResolve        bool `json:"can_resolve"`
	CanRequestReview  bool `json:"can_request_review"`
	CanReview         bool `json:"can_review"`
	CanFinalize       bool `json:"can_finalize"`
	CanReviseReturned bool `json:"can_revise_returned"`
}

type TimelineItem struct {
	Sequence   int64  `json:"sequence"`
	Kind       string `json:"kind"`
	ActorID    string `json:"actor_id"`
	OccurredAt string `json:"occurred_at"`
	Digest     string `json:"digest"`
}

type CaseView struct {
	Case       *domain.ReconstructionCase `json:"case"`
	Sherds     []domain.SherdRecord       `json:"sherds"`
	Hypotheses []domain.JoinHypothesis    `json:"hypotheses"`
	Challenges []domain.Challenge         `json:"challenges"`
	Actions    ActionSet                  `json:"actions"`
	Timeline   []TimelineItem             `json:"timeline"`
}

type CaseSummary struct {
	CaseID      string            `json:"case_id"`
	SiteUnit    string            `json:"site_unit"`
	VesselClass string            `json:"vessel_class"`
	OwnerID     string            `json:"owner_id"`
	Status      domain.CaseStatus `json:"status"`
	Revision    int64             `json:"revision"`
	UpdatedAt   string            `json:"updated_at"`
}

type DossierView struct {
	Dossier *domain.FinalDossier `json:"dossier"`
	Valid   bool                 `json:"valid"`
	Message string               `json:"message"`
}

func (s *Service) Detail(ctx context.Context, caseID string) (*CaseView, error) {
	c, err := s.store.Load(ctx, caseID)
	if err != nil {
		return nil, mapStoreError(err)
	}
	return s.caseView(ctx, c)
}

func (s *Service) ListCases(ctx context.Context) ([]CaseSummary, error) {
	cases, err := s.store.List(ctx, 100)
	if err != nil {
		return nil, err
	}
	result := make([]CaseSummary, 0, len(cases))
	for _, c := range cases {
		result = append(result, CaseSummary{CaseID: c.CaseID, SiteUnit: c.SiteUnit, VesselClass: c.VesselClass,
			OwnerID: c.OwnerID, Status: c.Status, Revision: c.Revision, UpdatedAt: c.UpdatedAt.UTC().Format("2006-01-02 15:04:05")})
	}
	return result, nil
}

func (s *Service) caseView(ctx context.Context, c *domain.ReconstructionCase) (*CaseView, error) {
	sherdIDs := make([]string, 0, len(c.Sherds))
	for id := range c.Sherds {
		sherdIDs = append(sherdIDs, id)
	}
	sort.Strings(sherdIDs)
	sherds := make([]domain.SherdRecord, 0, len(sherdIDs))
	for _, id := range sherdIDs {
		sherds = append(sherds, *c.Sherds[id])
	}
	hypothesisIDs := make([]string, 0, len(c.Hypotheses))
	for id := range c.Hypotheses {
		hypothesisIDs = append(hypothesisIDs, id)
	}
	sort.Strings(hypothesisIDs)
	hypotheses := make([]domain.JoinHypothesis, 0, len(hypothesisIDs))
	for _, id := range hypothesisIDs {
		hypotheses = append(hypotheses, *c.Hypotheses[id])
	}
	challengeIDs := make([]string, 0, len(c.Challenges))
	for id := range c.Challenges {
		challengeIDs = append(challengeIDs, id)
	}
	sort.Strings(challengeIDs)
	challenges := make([]domain.Challenge, 0, len(challengeIDs))
	for _, id := range challengeIDs {
		challenges = append(challenges, *c.Challenges[id])
	}
	events, err := s.store.Events(ctx, c.CaseID)
	if err != nil {
		return nil, err
	}
	timeline := make([]TimelineItem, 0, len(events))
	for _, event := range events {
		timeline = append(timeline, TimelineItem{Sequence: event.Sequence, Kind: event.Kind, ActorID: event.ActorID, OccurredAt: event.OccurredAt.UTC().Format("2006-01-02 15:04:05"), Digest: event.Digest})
	}
	actions := ActionSet{
		CanAddSherd: c.Status == domain.CaseDraft, CanFreezeBaseline: c.Status == domain.CaseDraft && len(c.Sherds) >= 2,
		CanAddHypothesis: c.Status == domain.CaseBaselineFrozen,
		CanChallenge:     c.Status == domain.CaseDeliberation, CanResolve: c.Status == domain.CaseDeliberation,
		CanRequestReview: c.Status == domain.CaseDeliberation, CanReview: c.Status == domain.CasePendingReview,
		CanFinalize: c.Status == domain.CaseApproved, CanReviseReturned: c.Status == domain.CaseChangesRequested,
	}
	return &CaseView{Case: c, Sherds: sherds, Hypotheses: hypotheses, Challenges: challenges, Actions: actions, Timeline: timeline}, nil
}

func dossierView(d *domain.FinalDossier, valid bool, message string) DossierView {
	return DossierView{Dossier: d, Valid: valid, Message: message}
}

func (s *Service) GetDossier(ctx context.Context, caseID string) (*DossierView, error) {
	dossier, err := s.store.Dossier(ctx, caseID)
	if err != nil {
		return nil, mapStoreError(err)
	}
	valid, message, err := s.store.VerifyDossier(ctx, caseID)
	if err != nil {
		return nil, mapStoreError(err)
	}
	view := dossierView(dossier, valid, message)
	return &view, nil
}

func (s *Service) VerifyDossier(ctx context.Context, caseID string) (*DossierView, error) {
	return s.GetDossier(ctx, caseID)
}
