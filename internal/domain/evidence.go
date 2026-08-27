package domain

import (
	"sort"
	"strings"
)

type EvidenceAssessment struct {
	Completeness int      `json:"completeness"`
	Missing      []string `json:"missing"`
}

func AssessEvidence(e Evidence) EvidenceAssessment {
	missing := make([]string, 0, len(EvidenceKeys))
	if strings.TrimSpace(e.EdgeMatch) == "" {
		missing = append(missing, "edge_match")
	}
	if strings.TrimSpace(e.FabricMatch) == "" {
		missing = append(missing, "fabric_match")
	}
	if strings.TrimSpace(e.DecorationContinuity) == "" {
		missing = append(missing, "decoration_continuity")
	}
	validScale := len(e.ScaleMeasurements) > 0
	for key, value := range e.ScaleMeasurements {
		if strings.TrimSpace(key) == "" || value <= 0 {
			validScale = false
		}
	}
	if !validScale {
		missing = append(missing, "scale_measurements")
	}
	validImages := len(e.ImageRefs) > 0
	for _, image := range e.ImageRefs {
		if strings.TrimSpace(image) == "" {
			validImages = false
		}
	}
	if !validImages {
		missing = append(missing, "image_refs")
	}
	sort.Strings(missing)
	return EvidenceAssessment{Completeness: (len(EvidenceKeys) - len(missing)) * 100 / len(EvidenceKeys), Missing: missing}
}

func CloneEvidence(e Evidence) Evidence {
	clone := e
	clone.ScaleMeasurements = make(map[string]float64, len(e.ScaleMeasurements))
	for key, value := range e.ScaleMeasurements {
		clone.ScaleMeasurements[key] = value
	}
	clone.ImageRefs = append([]string(nil), e.ImageRefs...)
	return clone
}

func EvidenceValue(e Evidence, key string) any {
	switch key {
	case "edge_match":
		return e.EdgeMatch
	case "fabric_match":
		return e.FabricMatch
	case "decoration_continuity":
		return e.DecorationContinuity
	case "scale_measurements":
		return e.ScaleMeasurements
	case "image_refs":
		return e.ImageRefs
	default:
		return nil
	}
}

func SetEvidenceValue(e *Evidence, key string, value any) error {
	if !ValidEvidenceKey(key) {
		return NewError(CodeValidation, "evidence_key", "未知证据项")
	}
	switch key {
	case "edge_match":
		text, ok := value.(string)
		if !ok || strings.TrimSpace(text) == "" {
			return NewError(CodeValidation, key, "补充内容不能为空")
		}
		e.EdgeMatch = text
	case "fabric_match":
		text, ok := value.(string)
		if !ok || strings.TrimSpace(text) == "" {
			return NewError(CodeValidation, key, "补充内容不能为空")
		}
		e.FabricMatch = text
	case "decoration_continuity":
		text, ok := value.(string)
		if !ok || strings.TrimSpace(text) == "" {
			return NewError(CodeValidation, key, "补充内容不能为空")
		}
		e.DecorationContinuity = text
	case "scale_measurements":
		measurements, ok := value.(map[string]float64)
		if !ok || len(measurements) == 0 {
			return NewError(CodeValidation, key, "补充测量不能为空")
		}
		e.ScaleMeasurements = measurements
	case "image_refs":
		images, ok := value.([]string)
		if !ok || len(images) == 0 {
			return NewError(CodeValidation, key, "补充图像引用不能为空")
		}
		e.ImageRefs = images
	}
	return nil
}
