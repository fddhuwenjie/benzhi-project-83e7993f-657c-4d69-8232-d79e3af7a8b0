package web

import (
	"net/http"

	"sherd-proof/internal/app"
)

func (a *API) RaiseChallengeHandler(w http.ResponseWriter, r *http.Request) {
	var command app.RaiseChallengeCommand
	if err := decodeJSON(w, r, &command); err != nil {
		writeError(w, err)
		return
	}
	command.CaseID = r.PathValue("caseID")
	result, err := a.service.RaiseChallenge(r.Context(), command)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *API) ResolveChallengeHandler(w http.ResponseWriter, r *http.Request) {
	var command app.ResolveChallengeCommand
	if err := decodeJSON(w, r, &command); err != nil {
		writeError(w, err)
		return
	}
	command.CaseID, command.ChallengeID = r.PathValue("caseID"), r.PathValue("challengeID")
	result, err := a.service.ResolveChallenge(r.Context(), command)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *API) RequestReviewHandler(w http.ResponseWriter, r *http.Request) {
	var command app.CaseCommand
	if err := decodeJSON(w, r, &command); err != nil {
		writeError(w, err)
		return
	}
	command.CaseID = r.PathValue("caseID")
	result, err := a.service.AdvanceToReview(r.Context(), command)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
