package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Deps struct {
	TenantSvc TenantService
}

type Server struct {
	r    chi.Router
	deps Deps
}

func New(deps Deps) *Server {
	r := chi.NewRouter()
	s := &Server{r: r, deps: deps}

	r.Get("/healthz", s.handleHealth)

	r.Route("/v1", func(r chi.Router) {
		r.Route("/tenants", func(r chi.Router) {
			r.Get("/", s.handleListTenants)
			r.Post("/", s.handleCreateTenant)
			r.Get("/{slug}", s.handleGetTenantBySlug)
		})
	})

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.r.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
