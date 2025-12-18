package httpapi_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gochatbot/internal/domain"
	"gochatbot/internal/httpapi"
	"gochatbot/internal/pagination"
)

type fakeTenantSvc struct {
	createErr error
	getErr    error
	listErr   error

	lastLimit  int
	lastCursor *pagination.Cursor
}

func (f *fakeTenantSvc) ListTenants(_ httpapi.RequestContext, limit int, cursor *pagination.Cursor) (httpapi.ListTenantsResult, error) {
	if f.listErr != nil {
		return httpapi.ListTenantsResult{}, f.listErr
	}
	f.lastLimit = limit
	f.lastCursor = cursor
	return httpapi.ListTenantsResult{
		Items: []httpapi.Tenant{
			{ID: "t1", Name: "Acme", Slug: "acme"},
		},
		NextCursor: "",
	}, nil
}

func (f *fakeTenantSvc) CreateTenant(_ httpapi.RequestContext, name, slug string) (httpapi.Tenant, error) {
	if f.createErr != nil {
		return httpapi.Tenant{}, f.createErr
	}
	return httpapi.Tenant{ID: "t1", Name: name, Slug: slug}, nil
}

func (f *fakeTenantSvc) GetTenantBySlug(_ httpapi.RequestContext, slug string) (httpapi.Tenant, error) {
	if f.getErr != nil {
		return httpapi.Tenant{}, f.getErr
	}
	return httpapi.Tenant{ID: "t1", Name: "Acme", Slug: slug}, nil
}

func TestHealthz(t *testing.T) {
	s := httpapi.New(httpapi.Deps{TenantSvc: &fakeTenantSvc{}})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/healthz", nil)

	s.ServeHTTP(rr, req)
	require.Equal(t, 200, rr.Code)
	require.Equal(t, "ok", rr.Body.String())
}

func TestCreateTenant_OK(t *testing.T) {
	fake := &fakeTenantSvc{}
	s := httpapi.New(httpapi.Deps{TenantSvc: fake})

	body := []byte(`{"name":"Acme Law","slug":"Acme Law!!"}`)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/tenants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	s.ServeHTTP(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)

	var got httpapi.Tenant
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	require.Equal(t, "Acme Law", got.Name)
	require.Equal(t, "acme-law", got.Slug)
}

func TestCreateTenant_InvalidSlug(t *testing.T) {
	s := httpapi.New(httpapi.Deps{TenantSvc: &fakeTenantSvc{}})

	body := []byte(`{"name":"Acme","slug":"!!"}`)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/tenants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	s.ServeHTTP(rr, req)
	require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

func TestCreateTenant_BadJSON(t *testing.T) {
	s := httpapi.New(httpapi.Deps{TenantSvc: &fakeTenantSvc{}})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/tenants", bytes.NewReader([]byte(`{nope`)))
	req.Header.Set("Content-Type", "application/json")

	s.ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetTenantBySlug_404(t *testing.T) {
	s := httpapi.New(httpapi.Deps{
		TenantSvc: &fakeTenantSvc{getErr: domain.ErrTenantNotFound},
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/tenants/acme-law", nil)

	s.ServeHTTP(rr, req)
	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetTenantBySlug_500(t *testing.T) {
	s := httpapi.New(httpapi.Deps{
		TenantSvc: &fakeTenantSvc{getErr: errors.New("db exploded")},
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/tenants/acme-law", nil)

	s.ServeHTTP(rr, req)
	require.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestCreateTenant_ConflictSlugTaken(t *testing.T) {
	s := httpapi.New(httpapi.Deps{
		TenantSvc: &fakeTenantSvc{createErr: domain.ErrTenantSlugTaken},
	})

	body := []byte(`{"name":"Acme","slug":"acme-law"}`)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/tenants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	s.ServeHTTP(rr, req)
	require.Equal(t, http.StatusConflict, rr.Code)
}

func TestListTenants_DefaultLimit(t *testing.T) {
	f := &fakeTenantSvc{}
	s := httpapi.New(httpapi.Deps{TenantSvc: f})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/tenants", nil)

	s.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, 50, f.lastLimit)
	require.Nil(t, f.lastCursor)
}

func TestListTenants_InvalidLimit(t *testing.T) {
	f := &fakeTenantSvc{}
	s := httpapi.New(httpapi.Deps{TenantSvc: f})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/tenants?limit=lol", nil)

	s.ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestListTenants_InvalidCursor(t *testing.T) {
	f := &fakeTenantSvc{}
	s := httpapi.New(httpapi.Deps{TenantSvc: f})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/tenants?cursor=not_base64", nil)

	s.ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestListTenants_ValidCursor(t *testing.T) {
	f := &fakeTenantSvc{}
	s := httpapi.New(httpapi.Deps{TenantSvc: f})

	// use real encoder to generate a valid cursor
	c := pagination.Cursor{CreatedAt: time.Date(2025, 12, 18, 0, 0, 0, 0, time.UTC), ID: "t9"}
	cur := pagination.Encode(c)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/tenants?cursor="+cur, nil)

	s.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, f.lastCursor)
	require.Equal(t, "t9", f.lastCursor.ID)
}
