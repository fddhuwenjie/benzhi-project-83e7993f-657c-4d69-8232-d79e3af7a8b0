package domain

import (
	"testing"
	"time"
)

func validSherd(id string) SherdRecord {
	return SherdRecord{SherdID: id, ContextCode: "T1H1", FabricCode: "F1", RimProfile: "断面完整",
		Dimensions: Dimensions{Height: 10, Width: 20, Depth: 5}, ImageRef: "archive://" + id,
		ImageDigest: "7c222fb2927d828af22f592134e8932480637c0d1c89b118e45e509fa60f4322"}
}

func completeEvidence() Evidence {
	return Evidence{EdgeMatch: "三点对应", FabricMatch: "胎土一致", DecorationContinuity: "弦纹连续",
		ScaleMeasurements: map[string]float64{"gap": 0.3}, ImageRefs: []string{"archive://overlay"}}
}

func reviewReadyCase(t *testing.T) *ReconstructionCase {
	t.Helper()
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	c, err := NewCase("C1", "T1", "灰陶罐", "owner", now)
	if err != nil {
		t.Fatal(err)
	}
	if err = c.AddSherd(validSherd("S1"), "owner", now); err != nil {
		t.Fatal(err)
	}
	if err = c.AddSherd(validSherd("S2"), "owner", now); err != nil {
		t.Fatal(err)
	}
	if err = c.FreezeBaseline("owner", now); err != nil {
		t.Fatal(err)
	}
	if err = c.AddHypothesis(JoinHypothesis{HypothesisID: "H1", SherdIDs: []string{"S1", "S2"}, Evidence: completeEvidence()}, "editor", now); err != nil {
		t.Fatal(err)
	}
	if err = c.SubmitHypothesis("H1", "editor", now); err != nil {
		t.Fatal(err)
	}
	if err = c.AdvanceToReview(now); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestLifecycleAndReviewerIndependence(t *testing.T) {
	c := reviewReadyCase(t)
	now := time.Now().UTC()
	if err := c.Review("editor", ReviewApprove, "同意", nil, now); !IsCode(err, CodeForbidden) {
		t.Fatalf("参与编辑者应被拒绝，得到 %v", err)
	}
	if err := c.Review("reviewer", ReviewApprove, "证据充分", nil, now); err != nil {
		t.Fatal(err)
	}
	if c.Status != CaseApproved {
		t.Fatalf("状态=%s", c.Status)
	}
	if err := c.AddSherd(validSherd("S3"), "owner", now); !IsCode(err, CodeState) {
		t.Fatalf("批准后应只读，得到 %v", err)
	}
}

func TestReturnedEvidenceOnlyAllowsSpecifiedKey(t *testing.T) {
	c := reviewReadyCase(t)
	now := time.Now().UTC()
	if err := c.Review("reviewer", ReviewReturn, "补充尺度", []string{"scale_measurements"}, now); err != nil {
		t.Fatal(err)
	}
	if err := c.ReviseReturnedEvidence("H1", "edge_match", "新描述", "editor", "修改", now); !IsCode(err, CodeForbidden) {
		t.Fatalf("未开放项应被拒绝，得到 %v", err)
	}
	if err := c.SubmitHypothesis("H1", "editor", now); !IsCode(err, CodeState) {
		t.Fatalf("未修订即提交应被拒绝，得到 %v", err)
	}
	if err := c.ReviseReturnedEvidence("H1", "scale_measurements", map[string]float64{"gap": 0.2}, "editor", "复测", now); err != nil {
		t.Fatal(err)
	}
	if c.Hypotheses["H1"].EvidenceVersion != 2 {
		t.Fatalf("证据版本未递增")
	}
	if err := c.SubmitHypothesis("H1", "editor", now); err != nil {
		t.Fatal(err)
	}
}
