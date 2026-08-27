package domain

import "testing"

func TestAssessEvidenceStableMissingKeys(t *testing.T) {
	assessment := AssessEvidence(Evidence{EdgeMatch: "存在", ScaleMeasurements: map[string]float64{"gap": 0}})
	if assessment.Completeness != 20 {
		t.Fatalf("完整度=%d", assessment.Completeness)
	}
	want := []string{"decoration_continuity", "fabric_match", "image_refs", "scale_measurements"}
	for index := range want {
		if assessment.Missing[index] != want[index] {
			t.Fatalf("缺项排序=%v", assessment.Missing)
		}
	}
}
