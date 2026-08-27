package web

import (
	"net/http"
)

func (a *API) WorkbenchHandler(w http.ResponseWriter, r *http.Request) {
	data, err := assets.ReadFile("web/index.html")
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (a *API) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if err := a.service.Store().Ping(r.Context()); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) ListCasesHandler(w http.ResponseWriter, r *http.Request) {
	result, err := a.service.ListCases(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"cases": result})
}

func (a *API) CaseDetailHandler(w http.ResponseWriter, r *http.Request) {
	result, err := a.service.Detail(r.Context(), r.PathValue("caseID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *API) DossierHandler(w http.ResponseWriter, r *http.Request) {
	result, err := a.service.GetDossier(r.Context(), r.PathValue("caseID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *API) VerifyDossierHandler(w http.ResponseWriter, r *http.Request) {
	result, err := a.service.VerifyDossier(r.Context(), r.PathValue("caseID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
