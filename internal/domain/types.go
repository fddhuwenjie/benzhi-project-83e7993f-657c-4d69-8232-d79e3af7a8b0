package domain

import "time"

type CaseStatus string

const (
	CaseDraft            CaseStatus = "draft"
	CaseBaselineFrozen   CaseStatus = "baseline_frozen"
	CaseDeliberation     CaseStatus = "deliberation"
	CasePendingReview    CaseStatus = "pending_review"
	CaseChangesRequested CaseStatus = "changes_requested"
	CaseApproved         CaseStatus = "approved"
	CaseSealed           CaseStatus = "sealed"
)

type HypothesisStatus string

const (
	HypothesisDraft     HypothesisStatus = "draft"
	HypothesisPublished HypothesisStatus = "published"
	HypothesisWithdrawn HypothesisStatus = "withdrawn"
	HypothesisApproved  HypothesisStatus = "approved"
)

type ChallengeStatus string

const (
	ChallengeOpen   ChallengeStatus = "open"
	ChallengeClosed ChallengeStatus = "closed"
)

type ResolutionKind string

const (
	ResolutionSupplement ResolutionKind = "supplement"
	ResolutionMaintain   ResolutionKind = "maintain"
	ResolutionWithdraw   ResolutionKind = "withdraw"
)

type ReviewDecision string

const (
	ReviewApprove ReviewDecision = "approve"
	ReviewReturn  ReviewDecision = "return"
)

var EvidenceKeys = []string{"edge_match", "fabric_match", "decoration_continuity", "scale_measurements", "image_refs"}

type Dimensions struct {
	Height float64 `json:"height"`
	Width  float64 `json:"width"`
	Depth  float64 `json:"depth"`
}

type SherdRecord struct {
	SherdID     string     `json:"sherd_id"`
	CaseID      string     `json:"case_id"`
	ContextCode string     `json:"context_code"`
	FabricCode  string     `json:"fabric_code"`
	RimProfile  string     `json:"rim_profile"`
	Dimensions  Dimensions `json:"dimensions_mm"`
	ImageRef    string     `json:"image_ref"`
	ImageDigest string     `json:"image_digest"`
}

type Evidence struct {
	EdgeMatch            string             `json:"edge_match"`
	FabricMatch          string             `json:"fabric_match"`
	DecorationContinuity string             `json:"decoration_continuity"`
	ScaleMeasurements    map[string]float64 `json:"scale_measurements"`
	ImageRefs            []string           `json:"image_refs"`
}

type EvidenceVersion struct {
	Version   int       `json:"version"`
	Evidence  Evidence  `json:"evidence"`
	ChangedBy string    `json:"changed_by"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

type JoinHypothesis struct {
	HypothesisID     string            `json:"hypothesis_id"`
	CaseID           string            `json:"case_id"`
	SherdIDs         []string          `json:"sherd_ids"`
	Evidence         Evidence          `json:"evidence"`
	EvidenceVersions []EvidenceVersion `json:"evidence_versions"`
	EvidenceVersion  int               `json:"evidence_version"`
	Completeness     int               `json:"completeness"`
	Missing          []string          `json:"missing"`
	Status           HypothesisStatus  `json:"status"`
	AuthorID         string            `json:"author_id"`
	LockedKeys       map[string]bool   `json:"locked_keys"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

type Challenge struct {
	ChallengeID    string          `json:"challenge_id"`
	HypothesisID   string          `json:"hypothesis_id"`
	EvidenceKey    string          `json:"evidence_key"`
	RaisedBy       string          `json:"raised_by"`
	Statement      string          `json:"statement"`
	Status         ChallengeStatus `json:"status"`
	ResolutionKind ResolutionKind  `json:"resolution_kind,omitempty"`
	ResolutionNote string          `json:"resolution_note,omitempty"`
	ResolvedBy     string          `json:"resolved_by,omitempty"`
	ResolvedAt     *time.Time      `json:"resolved_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type ReviewRecord struct {
	ReviewerID string         `json:"reviewer_id"`
	Decision   ReviewDecision `json:"decision"`
	Reason     string         `json:"reason"`
	ReopenKeys []string       `json:"reopen_keys,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

type ReconstructionCase struct {
	CaseID           string                     `json:"case_id"`
	SiteUnit         string                     `json:"site_unit"`
	VesselClass      string                     `json:"vessel_class"`
	OwnerID          string                     `json:"owner_id"`
	Status           CaseStatus                 `json:"status"`
	Revision         int64                      `json:"revision"`
	BaselineFrozenAt *time.Time                 `json:"baseline_frozen_at,omitempty"`
	CreatedAt        time.Time                  `json:"created_at"`
	UpdatedAt        time.Time                  `json:"updated_at"`
	Sherds           map[string]*SherdRecord    `json:"sherds"`
	Hypotheses       map[string]*JoinHypothesis `json:"hypotheses"`
	Challenges       map[string]*Challenge      `json:"challenges"`
	Reviews          []ReviewRecord             `json:"reviews"`
	Editors          map[string]bool            `json:"editors"`
	ReopenedKeys     map[string]bool            `json:"reopened_keys"`
}

type AuditEvent struct {
	Sequence   int64     `json:"sequence"`
	CaseID     string    `json:"case_id"`
	Kind       string    `json:"kind"`
	ActorID    string    `json:"actor_id"`
	OccurredAt time.Time `json:"occurred_at"`
	Payload    []byte    `json:"payload"`
	Previous   string    `json:"previous"`
	Digest     string    `json:"digest"`
}

type FinalDossier struct {
	DossierID             string    `json:"dossier_id"`
	CaseID                string    `json:"case_id"`
	ApprovedHypothesisIDs []string  `json:"approved_hypothesis_ids"`
	ReviewerID            string    `json:"reviewer_id"`
	ReviewReason          string    `json:"review_reason"`
	CanonicalPayload      []byte    `json:"canonical_payload"`
	EventChainHead        string    `json:"event_chain_head"`
	SHA256                string    `json:"sha256"`
	SealedAt              time.Time `json:"sealed_at"`
}
