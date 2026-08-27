package domain

import (
	"encoding/hex"
	"strings"
)

func ValidateCaseIdentity(siteUnit, vesselClass, ownerID string) error {
	if strings.TrimSpace(siteUnit) == "" {
		return NewError(CodeValidation, "site_unit", "遗址单元不能为空")
	}
	if strings.TrimSpace(vesselClass) == "" {
		return NewError(CodeValidation, "vessel_class", "器物类别不能为空")
	}
	if strings.TrimSpace(ownerID) == "" {
		return NewError(CodeValidation, "owner_id", "负责人不能为空")
	}
	return nil
}

func ValidateSherd(s SherdRecord) error {
	if strings.TrimSpace(s.SherdID) == "" {
		return NewError(CodeValidation, "sherd_id", "陶片编号不能为空")
	}
	if strings.TrimSpace(s.ContextCode) == "" {
		return NewError(CodeValidation, "context_code", "出土单位编码不能为空")
	}
	if strings.TrimSpace(s.FabricCode) == "" {
		return NewError(CodeValidation, "fabric_code", "胎土编码不能为空")
	}
	if strings.TrimSpace(s.RimProfile) == "" {
		return NewError(CodeValidation, "rim_profile", "断面描述不能为空")
	}
	if s.Dimensions.Height <= 0 || s.Dimensions.Width <= 0 || s.Dimensions.Depth <= 0 {
		return NewError(CodeValidation, "dimensions_mm", "三项尺寸必须大于零")
	}
	if strings.TrimSpace(s.ImageRef) == "" {
		return NewError(CodeValidation, "image_ref", "图像引用不能为空")
	}
	digest, err := hex.DecodeString(s.ImageDigest)
	if err != nil || len(digest) != 32 {
		return NewError(CodeValidation, "image_digest", "图像摘要必须是 64 位 SHA-256 十六进制值")
	}
	return nil
}

func ValidEvidenceKey(key string) bool {
	for _, candidate := range EvidenceKeys {
		if key == candidate {
			return true
		}
	}
	return false
}
