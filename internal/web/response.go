package web

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"sherd-proof/internal/domain"
)

const maxBodyBytes = 1 << 20

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		if errors.Is(err, io.EOF) {
			return domain.NewError(domain.CodeValidation, "body", "请求体不能为空")
		}
		return domain.NewError(domain.CodeValidation, "body", "JSON 请求格式不正确: "+err.Error())
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return domain.NewError(domain.CodeValidation, "body", "请求体只能包含一个 JSON 对象")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	body := errorBody{Code: "internal", Message: "服务处理失败"}
	var rule *domain.RuleError
	if errors.As(err, &rule) {
		body = errorBody{Code: string(rule.Code), Field: rule.Field, Message: rule.Message}
		switch rule.Code {
		case domain.CodeValidation:
			status = http.StatusBadRequest
		case domain.CodeNotFound:
			status = http.StatusNotFound
		case domain.CodeConflict, domain.CodeState:
			status = http.StatusConflict
		case domain.CodeForbidden:
			status = http.StatusForbidden
		}
	}
	writeJSON(w, status, errorResponse{Error: body})
}
