package web

import (
	"context"
	"embed"
	"io/fs"
	"net/http"
	"time"

	"sherd-proof/internal/app"
)

//go:embed web/*
var assets embed.FS

type API struct {
	service *app.Service
	mux     *http.ServeMux
}

func New(service *app.Service) http.Handler {
	api := &API{service: service, mux: http.NewServeMux()}
	api.routes()
	return securityHeaders(api.mux)
}

func (a *API) routes() {
	static, _ := fs.Sub(assets, "web")
	a.mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(static))))
	a.mux.HandleFunc("GET /", a.WorkbenchHandler)
	a.mux.HandleFunc("GET /healthz", a.HealthHandler)
	a.mux.HandleFunc("GET /api/cases", a.ListCasesHandler)
	a.mux.HandleFunc("POST /api/cases", a.CreateCaseHandler)
	a.mux.HandleFunc("GET /api/cases/{caseID}", a.CaseDetailHandler)
	a.mux.HandleFunc("POST /api/cases/{caseID}/sherds", a.AddSherdHandler)
	a.mux.HandleFunc("POST /api/cases/{caseID}/freeze", a.FreezeBaselineHandler)
	a.mux.HandleFunc("POST /api/cases/{caseID}/hypotheses", a.AddHypothesisHandler)
	a.mux.HandleFunc("POST /api/cases/{caseID}/hypotheses/{hypothesisID}/submit", a.SubmitHypothesisHandler)
	a.mux.HandleFunc("POST /api/cases/{caseID}/hypotheses/{hypothesisID}/evidence/{evidenceKey}/revise", a.ReviseEvidenceHandler)
	a.mux.HandleFunc("POST /api/cases/{caseID}/challenges", a.RaiseChallengeHandler)
	a.mux.HandleFunc("POST /api/cases/{caseID}/challenges/{challengeID}/resolve", a.ResolveChallengeHandler)
	a.mux.HandleFunc("POST /api/cases/{caseID}/request-review", a.RequestReviewHandler)
	a.mux.HandleFunc("POST /api/cases/{caseID}/review", a.ReviewHandler)
	a.mux.HandleFunc("POST /api/cases/{caseID}/finalize", a.FinalizeHandler)
	a.mux.HandleFunc("GET /api/cases/{caseID}/dossier", a.DossierHandler)
	a.mux.HandleFunc("GET /api/cases/{caseID}/dossier/verify", a.VerifyDossierHandler)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
