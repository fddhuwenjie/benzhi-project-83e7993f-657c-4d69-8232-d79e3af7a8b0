package domain

import (
	"sort"
	"strings"
	"time"
)

func NewCase(id, siteUnit, vesselClass, ownerID string, now time.Time) (*ReconstructionCase, error) {
	if strings.TrimSpace(id) == "" {
		return nil, NewError(CodeValidation, "case_id", "案件编号不能为空")
	}
	if err := ValidateCaseIdentity(siteUnit, vesselClass, ownerID); err != nil {
		return nil, err
	}
	return &ReconstructionCase{
		CaseID: id, SiteUnit: strings.TrimSpace(siteUnit), VesselClass: strings.TrimSpace(vesselClass),
		OwnerID: strings.TrimSpace(ownerID), Status: CaseDraft, Revision: 0, CreatedAt: now, UpdatedAt: now,
		Sherds: map[string]*SherdRecord{}, Hypotheses: map[string]*JoinHypothesis{},
		Challenges: map[string]*Challenge{}, Editors: map[string]bool{ownerID: true}, ReopenedKeys: map[string]bool{},
	}, nil
}

func (c *ReconstructionCase) ensureMutable() error {
	if c.Status == CaseSealed || c.Status == CaseApproved {
		return NewError(CodeState, "status", "案件已批准或定稿，不可修改")
	}
	return nil
}

func (c *ReconstructionCase) AddSherd(s SherdRecord, actor string, now time.Time) error {
	if err := c.ensureMutable(); err != nil {
		return err
	}
	if c.Status != CaseDraft {
		return NewError(CodeState, "status", "仅草稿阶段可登记陶片")
	}
	if err := ValidateSherd(s); err != nil {
		return err
	}
	if _, exists := c.Sherds[s.SherdID]; exists {
		return NewError(CodeConflict, "sherd_id", "陶片编号已存在")
	}
	s.CaseID = c.CaseID
	c.Sherds[s.SherdID] = &s
	c.Editors[actor] = true
	c.UpdatedAt = now
	return nil
}

func (c *ReconstructionCase) RemoveSherd(id, actor string, now time.Time) error {
	if c.Status != CaseDraft {
		return NewError(CodeState, "status", "基线冻结后不可增删陶片")
	}
	if _, exists := c.Sherds[id]; !exists {
		return NewError(CodeNotFound, "sherd_id", "陶片不存在")
	}
	delete(c.Sherds, id)
	c.Editors[actor] = true
	c.UpdatedAt = now
	return nil
}

func (c *ReconstructionCase) FreezeBaseline(actor string, now time.Time) error {
	if c.Status != CaseDraft {
		return NewError(CodeState, "status", "当前阶段不能冻结基线")
	}
	if len(c.Sherds) < 2 {
		return NewError(CodeValidation, "sherds", "冻结基线至少需要两件完整陶片")
	}
	for _, sherd := range c.Sherds {
		if err := ValidateSherd(*sherd); err != nil {
			return err
		}
	}
	c.Status = CaseBaselineFrozen
	c.BaselineFrozenAt = &now
	c.Editors[actor] = true
	c.UpdatedAt = now
	return nil
}

func (c *ReconstructionCase) AddHypothesis(h JoinHypothesis, actor string, now time.Time) error {
	if err := c.ensureMutable(); err != nil {
		return err
	}
	if c.Status != CaseBaselineFrozen {
		return NewError(CodeState, "status", "当前阶段不能登记候选")
	}
	if strings.TrimSpace(h.HypothesisID) == "" {
		return NewError(CodeValidation, "hypothesis_id", "候选编号不能为空")
	}
	if _, exists := c.Hypotheses[h.HypothesisID]; exists {
		return NewError(CodeConflict, "hypothesis_id", "候选编号已存在")
	}
	if len(h.SherdIDs) < 2 {
		return NewError(CodeValidation, "sherd_ids", "候选至少关联两件陶片")
	}
	seen := map[string]bool{}
	for _, id := range h.SherdIDs {
		if _, ok := c.Sherds[id]; !ok {
			return NewError(CodeValidation, "sherd_ids", "候选引用了基线外陶片")
		}
		if seen[id] {
			return NewError(CodeValidation, "sherd_ids", "候选陶片编号重复")
		}
		seen[id] = true
	}
	sort.Strings(h.SherdIDs)
	assessment := AssessEvidence(h.Evidence)
	h.CaseID, h.AuthorID, h.Completeness, h.Missing = c.CaseID, actor, assessment.Completeness, assessment.Missing
	h.Status, h.EvidenceVersion, h.CreatedAt, h.UpdatedAt = HypothesisDraft, 1, now, now
	h.LockedKeys = map[string]bool{}
	h.EvidenceVersions = []EvidenceVersion{{Version: 1, Evidence: CloneEvidence(h.Evidence), ChangedBy: actor, Note: "初始证据", CreatedAt: now}}
	c.Hypotheses[h.HypothesisID] = &h
	c.Editors[actor] = true
	c.UpdatedAt = now
	return nil
}

func (c *ReconstructionCase) SubmitHypothesis(id, actor string, now time.Time) error {
	if c.Status != CaseBaselineFrozen && c.Status != CaseChangesRequested {
		return NewError(CodeState, "status", "当前阶段不能提交候选")
	}
	h, ok := c.Hypotheses[id]
	if !ok {
		return NewError(CodeNotFound, "hypothesis_id", "候选不存在")
	}
	if h.AuthorID != actor {
		return NewError(CodeForbidden, "actor_id", "仅候选作者可提交")
	}
	if c.Status == CaseChangesRequested && len(c.ReopenedKeys) > 0 {
		return NewError(CodeState, "reopen_keys", "复核指定的证据项尚未全部修订")
	}
	assessment := AssessEvidence(h.Evidence)
	if assessment.Completeness != 100 {
		return NewError(CodeValidation, "evidence", "候选证据不完整")
	}
	h.Completeness, h.Missing, h.Status, h.UpdatedAt = assessment.Completeness, assessment.Missing, HypothesisPublished, now
	c.Status, c.UpdatedAt = CaseDeliberation, now
	return nil
}

func (c *ReconstructionCase) ReviseReturnedEvidence(hypothesisID, key string, replacement any, actor, note string, now time.Time) error {
	if c.Status != CaseChangesRequested {
		return NewError(CodeState, "status", "仅定向退回阶段可修订证据")
	}
	if !c.ReopenedKeys[key] {
		return NewError(CodeForbidden, "evidence_key", "该证据项未被复核开放")
	}
	h, ok := c.Hypotheses[hypothesisID]
	if !ok || h.Status == HypothesisWithdrawn {
		return NewError(CodeNotFound, "hypothesis_id", "有效候选不存在")
	}
	if h.AuthorID != actor {
		return NewError(CodeForbidden, "actor_id", "仅候选作者可修订退回证据")
	}
	if strings.TrimSpace(note) == "" {
		return NewError(CodeValidation, "note", "修订说明不能为空")
	}
	if err := SetEvidenceValue(&h.Evidence, key, replacement); err != nil {
		return err
	}
	h.EvidenceVersion++
	h.EvidenceVersions = append(h.EvidenceVersions, EvidenceVersion{Version: h.EvidenceVersion, Evidence: CloneEvidence(h.Evidence), ChangedBy: actor, Note: note, CreatedAt: now})
	assessment := AssessEvidence(h.Evidence)
	h.Completeness, h.Missing, h.Status, h.UpdatedAt = assessment.Completeness, assessment.Missing, HypothesisDraft, now
	delete(c.ReopenedKeys, key)
	c.Editors[actor] = true
	c.UpdatedAt = now
	return nil
}

func (c *ReconstructionCase) RaiseChallenge(ch Challenge, now time.Time) error {
	if c.Status != CaseDeliberation {
		return NewError(CodeState, "status", "仅公开评议阶段可提出异议")
	}
	if strings.TrimSpace(ch.ChallengeID) == "" || strings.TrimSpace(ch.Statement) == "" || strings.TrimSpace(ch.RaisedBy) == "" {
		return NewError(CodeValidation, "challenge", "异议编号、提出人和说明不能为空")
	}
	if !ValidEvidenceKey(ch.EvidenceKey) {
		return NewError(CodeValidation, "evidence_key", "未知证据项")
	}
	if _, exists := c.Challenges[ch.ChallengeID]; exists {
		return NewError(CodeConflict, "challenge_id", "异议编号已存在")
	}
	h, ok := c.Hypotheses[ch.HypothesisID]
	if !ok || h.Status != HypothesisPublished {
		return NewError(CodeState, "hypothesis_id", "只能质疑已公开的有效候选")
	}
	ch.Status, ch.CreatedAt = ChallengeOpen, now
	c.Challenges[ch.ChallengeID] = &ch
	h.LockedKeys[ch.EvidenceKey] = true
	c.UpdatedAt = now
	return nil
}

func (c *ReconstructionCase) ResolveChallenge(id string, kind ResolutionKind, note, actor string, replacement any, now time.Time) error {
	if c.Status != CaseDeliberation {
		return NewError(CodeState, "status", "当前阶段不能处置异议")
	}
	ch, ok := c.Challenges[id]
	if !ok {
		return NewError(CodeNotFound, "challenge_id", "异议不存在")
	}
	if ch.Status == ChallengeClosed {
		return NewError(CodeConflict, "challenge_id", "异议已关闭")
	}
	if strings.TrimSpace(note) == "" || strings.TrimSpace(actor) == "" {
		return NewError(CodeValidation, "resolution_note", "处置人和处置说明不能为空")
	}
	h := c.Hypotheses[ch.HypothesisID]
	switch kind {
	case ResolutionSupplement:
		if err := SetEvidenceValue(&h.Evidence, ch.EvidenceKey, replacement); err != nil {
			return err
		}
		h.EvidenceVersion++
		h.EvidenceVersions = append(h.EvidenceVersions, EvidenceVersion{Version: h.EvidenceVersion, Evidence: CloneEvidence(h.Evidence), ChangedBy: actor, Note: note, CreatedAt: now})
		assessment := AssessEvidence(h.Evidence)
		h.Completeness, h.Missing = assessment.Completeness, assessment.Missing
		c.Editors[actor] = true
	case ResolutionMaintain:
	case ResolutionWithdraw:
		h.Status = HypothesisWithdrawn
	default:
		return NewError(CodeValidation, "resolution_kind", "未知处置动作")
	}
	ch.Status, ch.ResolutionKind, ch.ResolutionNote, ch.ResolvedBy = ChallengeClosed, kind, note, actor
	ch.ResolvedAt, h.UpdatedAt, c.UpdatedAt = &now, now, now
	delete(h.LockedKeys, ch.EvidenceKey)
	return nil
}

func (c *ReconstructionCase) AdvanceToReview(now time.Time) error {
	if c.Status != CaseDeliberation {
		return NewError(CodeState, "status", "仅评议阶段可进入复核")
	}
	for _, challenge := range c.Challenges {
		if challenge.Status != ChallengeClosed {
			return NewError(CodeState, "challenges", "仍有未关闭异议")
		}
	}
	active := 0
	for _, hypothesis := range c.Hypotheses {
		if hypothesis.Status == HypothesisPublished {
			active++
		}
	}
	if active == 0 {
		return NewError(CodeState, "hypotheses", "至少需要一个有效候选")
	}
	c.Status, c.UpdatedAt = CasePendingReview, now
	return nil
}

func (c *ReconstructionCase) Review(reviewer string, decision ReviewDecision, reason string, reopenKeys []string, now time.Time) error {
	if c.Status != CasePendingReview {
		return NewError(CodeState, "status", "案件尚未进入待复核")
	}
	if strings.TrimSpace(reason) == "" {
		return NewError(CodeValidation, "reason", "复核理由不能为空")
	}
	if reviewer == c.OwnerID || c.Editors[reviewer] {
		return NewError(CodeForbidden, "reviewer_id", "复核员必须未参与建档和候选编辑")
	}
	record := ReviewRecord{ReviewerID: reviewer, Decision: decision, Reason: reason, CreatedAt: now}
	switch decision {
	case ReviewApprove:
		c.Status = CaseApproved
		for _, h := range c.Hypotheses {
			if h.Status == HypothesisPublished {
				h.Status = HypothesisApproved
			}
		}
	case ReviewReturn:
		if len(reopenKeys) == 0 {
			return NewError(CodeValidation, "reopen_keys", "退回必须指定开放的证据项")
		}
		c.ReopenedKeys = map[string]bool{}
		for _, key := range reopenKeys {
			if !ValidEvidenceKey(key) {
				return NewError(CodeValidation, "reopen_keys", "包含未知证据项")
			}
			c.ReopenedKeys[key] = true
		}
		record.ReopenKeys = append([]string(nil), reopenKeys...)
		c.Status = CaseChangesRequested
	default:
		return NewError(CodeValidation, "decision", "未知复核结论")
	}
	c.Reviews = append(c.Reviews, record)
	c.UpdatedAt = now
	return nil
}

func (c *ReconstructionCase) Seal(now time.Time) error {
	if c.Status != CaseApproved {
		return NewError(CodeState, "status", "仅复核通过案件可定稿")
	}
	c.Status, c.UpdatedAt = CaseSealed, now
	return nil
}
