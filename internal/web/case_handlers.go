package web

import (
	"net/http"

	"sherd-proof/internal/app"
)

func (a *API) CreateCaseHandler(w http.ResponseWriter, r *http.Request) {
	var command app.CreateCaseCommand
	if err := decodeJSON(w, r, &command); err != nil {
		writeError(w, err)
		return
	}
	result, err := a.service.CreateCase(r.Context(), command)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (a *API) AddSherdHandler(w http.ResponseWriter, r *http.Request) {
	var command app.AddSherdCommand
	if err := decodeJSON(w, r, &command); err != nil {
		writeError(w, err)
		return
	}
	command.CaseID = r.PathValue("caseID")
	result, err := a.service.AddSherd(r.Context(), command)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *API) FreezeBaselineHandler(w http.ResponseWriter, r *http.Request) {
	var command app.CaseCommand
	if err := decodeJSON(w, r, &command); err != nil {
		writeError(w, err)
		return
	}
	command.CaseID = r.PathValue("caseID")
	result, err := a.service.FreezeBaseline(r.Context(), command)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
