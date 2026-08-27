package app

import (
	"encoding/json"

	"sherd-proof/internal/domain"
)

type RequestMeta struct {
	RequestID        string `json:"request_id"`
	ExpectedRevision int64  `json:"expected_revision"`
}

type CreateCaseCommand struct {
	RequestMeta
	CaseID      string `json:"case_id"`
	SiteUnit    string `json:"site_unit"`
	VesselClass string `json:"vessel_class"`
	OwnerID     string `json:"owner_id"`
}

type AddSherdCommand struct {
	RequestMeta
	CaseID  string             `json:"case_id"`
	ActorID string             `json:"actor_id"`
	Sherd   domain.SherdRecord `json:"sherd"`
}

type CaseCommand struct {
	RequestMeta
	CaseID  string `json:"case_id"`
	ActorID string `json:"actor_id"`
}

type AddHypothesisCommand struct {
	RequestMeta
	CaseID       string          `json:"case_id"`
	ActorID      string          `json:"actor_id"`
	HypothesisID string          `json:"hypothesis_id"`
	SherdIDs     []string        `json:"sherd_ids"`
	Evidence     domain.Evidence `json:"evidence"`
}

type HypothesisCommand struct {
	RequestMeta
	CaseID       string `json:"case_id"`
	ActorID      string `json:"actor_id"`
	HypothesisID string `json:"hypothesis_id"`
}

type ReviseEvidenceCommand struct {
	RequestMeta
	CaseID       string          `json:"case_id"`
	ActorID      string          `json:"actor_id"`
	HypothesisID string          `json:"hypothesis_id"`
	EvidenceKey  string          `json:"evidence_key"`
	Note         string          `json:"note"`
	Replacement  json.RawMessage `json:"replacement"`
}

type RaiseChallengeCommand struct {
	RequestMeta
	CaseID       string `json:"case_id"`
	ActorID      string `json:"actor_id"`
	ChallengeID  string `json:"challenge_id"`
	HypothesisID string `json:"hypothesis_id"`
	EvidenceKey  string `json:"evidence_key"`
	Statement    string `json:"statement"`
}

type ResolveChallengeCommand struct {
	RequestMeta
	CaseID      string                `json:"case_id"`
	ActorID     string                `json:"actor_id"`
	ChallengeID string                `json:"challenge_id"`
	Kind        domain.ResolutionKind `json:"resolution_kind"`
	Note        string                `json:"resolution_note"`
	Replacement json.RawMessage       `json:"replacement,omitempty"`
}

type ReviewCommand struct {
	RequestMeta
	CaseID     string                `json:"case_id"`
	ReviewerID string                `json:"reviewer_id"`
	Decision   domain.ReviewDecision `json:"decision"`
	Reason     string                `json:"reason"`
	ReopenKeys []string              `json:"reopen_keys,omitempty"`
}

type FinalizeCommand struct {
	RequestMeta
	CaseID    string `json:"case_id"`
	ActorID   string `json:"actor_id"`
	DossierID string `json:"dossier_id"`
}
