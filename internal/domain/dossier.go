package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"time"
)

type dossierPayload struct {
	CaseID      string           `json:"case_id"`
	SiteUnit    string           `json:"site_unit"`
	VesselClass string           `json:"vessel_class"`
	Sherds      []SherdRecord    `json:"sherds"`
	Hypotheses  []JoinHypothesis `json:"hypotheses"`
	Challenges  []Challenge      `json:"challenges"`
	Review      ReviewRecord     `json:"review"`
	ChainHead   string           `json:"event_chain_head"`
	SealedAt    string           `json:"sealed_at"`
}

func BuildDossier(c *ReconstructionCase, dossierID, chainHead string, sealedAt time.Time) (*FinalDossier, error) {
	if c.Status != CaseApproved {
		return nil, NewError(CodeState, "status", "案件未获复核通过")
	}
	if len(c.Reviews) == 0 || c.Reviews[len(c.Reviews)-1].Decision != ReviewApprove {
		return nil, NewError(CodeState, "review", "缺少有效的通过结论")
	}
	sherdIDs := make([]string, 0, len(c.Sherds))
	for id := range c.Sherds {
		sherdIDs = append(sherdIDs, id)
	}
	sort.Strings(sherdIDs)
	sherds := make([]SherdRecord, 0, len(sherdIDs))
	for _, id := range sherdIDs {
		sherds = append(sherds, *c.Sherds[id])
	}
	hypothesisIDs := make([]string, 0)
	for id, h := range c.Hypotheses {
		if h.Status == HypothesisApproved {
			hypothesisIDs = append(hypothesisIDs, id)
		}
	}
	sort.Strings(hypothesisIDs)
	hypotheses := make([]JoinHypothesis, 0, len(hypothesisIDs))
	for _, id := range hypothesisIDs {
		hypotheses = append(hypotheses, *c.Hypotheses[id])
	}
	challengeIDs := make([]string, 0, len(c.Challenges))
	for id := range c.Challenges {
		challengeIDs = append(challengeIDs, id)
	}
	sort.Strings(challengeIDs)
	challenges := make([]Challenge, 0, len(challengeIDs))
	for _, id := range challengeIDs {
		challenges = append(challenges, *c.Challenges[id])
	}
	review := c.Reviews[len(c.Reviews)-1]
	payload := dossierPayload{CaseID: c.CaseID, SiteUnit: c.SiteUnit, VesselClass: c.VesselClass, Sherds: sherds,
		Hypotheses: hypotheses, Challenges: challenges, Review: review, ChainHead: chainHead, SealedAt: sealedAt.UTC().Format(time.RFC3339Nano)}
	canonical, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(canonical)
	return &FinalDossier{DossierID: dossierID, CaseID: c.CaseID, ApprovedHypothesisIDs: hypothesisIDs,
		ReviewerID: review.ReviewerID, ReviewReason: review.Reason, CanonicalPayload: canonical,
		EventChainHead: chainHead, SHA256: hex.EncodeToString(sum[:]), SealedAt: sealedAt}, nil
}

func VerifyDossier(d *FinalDossier, currentChainHead string) (bool, string) {
	if d == nil {
		return false, "档案不存在"
	}
	sum := sha256.Sum256(d.CanonicalPayload)
	if hex.EncodeToString(sum[:]) != d.SHA256 {
		return false, "规范化载荷校验失败"
	}
	if currentChainHead != d.EventChainHead {
		return false, "事件摘要链与定稿时不一致"
	}
	var payload dossierPayload
	if err := json.Unmarshal(d.CanonicalPayload, &payload); err != nil {
		return false, "规范化载荷无法解析"
	}
	if payload.CaseID != d.CaseID || payload.ChainHead != d.EventChainHead {
		return false, "档案身份或摘要链头不一致"
	}
	return true, "档案载荷与事件摘要链校验通过"
}
