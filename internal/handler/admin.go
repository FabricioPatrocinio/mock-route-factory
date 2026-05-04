package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	// TODO: replace SEU_USUARIO with your GitHub username
	"github.com/SEU_USUARIO/mock-route-factory/internal/model"
	"github.com/SEU_USUARIO/mock-route-factory/internal/repo"
)

var allowedMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodPost:    true,
	http.MethodPut:     true,
	http.MethodPatch:   true,
	http.MethodDelete:  true,
	http.MethodHead:    true,
	http.MethodOptions: true,
}

type Admin struct {
	repo          *repo.Repo
	publicBaseURL string
}

func NewAdmin(r *repo.Repo, publicBaseURL string) *Admin {
	return &Admin{repo: r, publicBaseURL: publicBaseURL}
}

type upsertRequest struct {
	Method string          `json:"method"`
	Path   string          `json:"path"`
	Status int             `json:"status"`
	Body   json.RawMessage `json:"response"`
}

func (a *Admin) Upsert(w http.ResponseWriter, r *http.Request) {
	var req upsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.Method = strings.ToUpper(strings.TrimSpace(req.Method))
	req.Path = normalizePath(req.Path)

	if !allowedMethods[req.Method] {
		jsonError(w, http.StatusUnprocessableEntity,
			fmt.Sprintf("method %q not allowed; accepted: GET POST PUT PATCH DELETE HEAD OPTIONS", req.Method))
		return
	}
	if len(req.Path) == 0 || req.Path[0] != '/' {
		jsonError(w, http.StatusUnprocessableEntity, "path must be non-empty and start with /")
		return
	}
	if isReserved(req.Path) {
		jsonError(w, http.StatusUnprocessableEntity,
			fmt.Sprintf("path %q is reserved and cannot be mocked", req.Path))
		return
	}
	if req.Status == 0 {
		req.Status = http.StatusOK
	}
	if req.Status < 100 || req.Status > 599 {
		jsonError(w, http.StatusUnprocessableEntity, "status must be between 100 and 599")
		return
	}
	if len(req.Body) == 0 || !json.Valid(req.Body) {
		jsonError(w, http.StatusUnprocessableEntity, "response is required and must be a valid JSON value")
		return
	}

	m, err := a.repo.Upsert(req.Method, req.Path, req.Status, []byte(req.Body))
	if err != nil {
		log.Printf("upsert error: %v", err)
		jsonError(w, http.StatusInternalServerError, "failed to save mock")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

func (a *Admin) List(w http.ResponseWriter, r *http.Request) {
	mocks, err := a.repo.List()
	if err != nil {
		log.Printf("list error: %v", err)
		jsonError(w, http.StatusInternalServerError, "failed to retrieve mocks")
		return
	}
	if mocks == nil {
		mocks = []*model.Mock{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mocks)
}

type curlItem struct {
	Mock *model.Mock `json:"mock"`
	Curl string      `json:"curl"`
}

func (a *Admin) Curls(w http.ResponseWriter, r *http.Request) {
	mocks, err := a.repo.List()
	if err != nil {
		log.Printf("curls list error: %v", err)
		jsonError(w, http.StatusInternalServerError, "failed to retrieve mocks")
		return
	}

	base := a.publicBaseURL
	if base == "" {
		base = deriveBaseURL(r)
	}

	items := make([]curlItem, 0, len(mocks))
	for _, m := range mocks {
		items = append(items, curlItem{Mock: m, Curl: buildCurl(base, m)})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (a *Admin) Delete(w http.ResponseWriter, r *http.Request) {
	method := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("method")))
	path := normalizePath(r.URL.Query().Get("path"))

	if method == "" || path == "" {
		jsonError(w, http.StatusBadRequest, "query parameters 'method' and 'path' are required")
		return
	}

	if err := a.repo.Delete(method, path); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			jsonError(w, http.StatusNotFound, "mock not found")
			return
		}
		log.Printf("delete error: %v", err)
		jsonError(w, http.StatusInternalServerError, "failed to delete mock")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// normalizePath trims whitespace and removes trailing slashes (root "/" is preserved).
func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	if p != "/" {
		p = strings.TrimRight(p, "/")
	}
	return p
}

// isReserved blocks /health (exact) and any path under /admin (prefix).
func isReserved(path string) bool {
	return path == "/health" || strings.HasPrefix(path, "/admin")
}

func deriveBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	return scheme + "://" + r.Host
}

func buildCurl(baseURL string, m *model.Mock) string {
	return fmt.Sprintf("curl -s -X %s '%s%s'", m.Method, baseURL, m.Path)
}

func jsonError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
