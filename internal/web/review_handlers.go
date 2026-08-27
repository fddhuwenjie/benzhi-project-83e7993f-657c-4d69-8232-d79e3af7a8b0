package web

import (
	"net/http"

	"sherd-proof/internal/app"
)

func (a *API) ReviewHandler(w http.ResponseWriter, r *http.Request) {
	var command app.ReviewCommand
	if err := decodeJSON(w, r, &command); err != nil {
		writeError(w, err)
		return
	}
	command.CaseID = r.PathValue("caseID")
	result, err := a.service.Review(r.Context(), command)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *API) FinalizeHandler(w http.ResponseWriter, r *http.Request) {
	var command app.FinalizeCommand
	if err := decodeJSON(w, r, &command); err != nil {
		writeError(w, err)
		return
	}
	command.CaseID = r.PathValue("caseID")
	result, err := a.service.Finalize(r.Context(), command)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
