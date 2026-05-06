package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/FabricioPatrocinio/mock-route-factory/internal/repo"
)

type Dynamic struct {
	repo *repo.Repo
}

func NewDynamic(r *repo.Repo) *Dynamic {
	return &Dynamic{repo: r}
}

func (d *Dynamic) Handle(w http.ResponseWriter, r *http.Request) {
	m, err := d.repo.Get(r.Method, r.URL.Path)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			jsonError(w, http.StatusNotFound,
				"no mock registered for "+r.Method+" "+r.URL.Path)
			return
		}
		log.Printf("dynamic handler error: %v", err)
		jsonError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(m.Status)
	_, _ = w.Write(m.ResponseBody)
}
