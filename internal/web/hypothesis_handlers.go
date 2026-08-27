package web

import (
	"net/http"

	"sherd-proof/internal/app"
)

func (a *API) AddHypothesisHandler(w http.ResponseWriter, r *http.Request) {
	var command app.AddHypothesisCommand
	if err := decodeJSON(w, r, &command); err != nil {
		writeError(w, err)
		return
	}
	command.CaseID = r.PathValue("caseID")
	result, err := a.service.AddHypothesis(r.Context(), command)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *API) SubmitHypothesisHandler(w http.ResponseWriter, r *http.Request) {
	var command app.HypothesisCommand
	if err := decodeJSON(w, r, &command); err != nil {
		writeError(w, err)
		return
	}
	command.CaseID, command.HypothesisID = r.PathValue("caseID"), r.PathValue("hypothesisID")
	result, err := a.service.SubmitHypothesis(r.Context(), command)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *API) ReviseEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	var command app.ReviseEvidenceCommand
	if err := decodeJSON(w, r, &command); err != nil {
		writeError(w, err)
		return
	}
	command.CaseID, command.HypothesisID, command.EvidenceKey = r.PathValue("caseID"), r.PathValue("hypothesisID"), r.PathValue("evidenceKey")
	result, err := a.service.ReviseReturnedEvidence(r.Context(), command)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
