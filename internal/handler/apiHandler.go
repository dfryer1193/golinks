package handler

import (
	"fmt"
	"github.com/dfryer1193/golinks/internal/links"
	"github.com/dfryer1193/golinks/internal/search"
	"github.com/dfryer1193/golinks/models"
	"github.com/dfryer1193/mjolnir/middleware"
	"github.com/dfryer1193/mjolnir/utils"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/url"
)

type alfredItem struct {
	Uid          string `json:"uid"`
	ObjType      string `json:"type"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	Arg          string `json:"arg"`
	Autocomplete string `json:"autocomplete"`
}

type alfredResponse struct {
	Items []alfredItem `json:"items"`
}

type ApiHandler struct {
	linkMap *links.LinkMap
}

func NewApiHandler(linkMap *links.LinkMap) *ApiHandler {
	return &ApiHandler{linkMap: linkMap}
}

func (h *ApiHandler) postLink(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "path")
	target := &struct {
		Target string `json:"target"`
	}{}
	err := utils.DecodeJSON(r, target)
	if err != nil {
		middleware.SetError(r, http.StatusBadRequest, fmt.Errorf("invalid target: %w", err))
		return
	}
	newEntry := &models.Entry{
		Path:   path,
		Target: target.Target,
	}

	targetUrl, err := url.Parse(target.Target)
	if err != nil {
		middleware.SetError(r, http.StatusBadRequest, fmt.Errorf("target %s is not a valid url", target.Target))
		return
	}

	var oldEntry *models.Entry
	oldTarget, exists := h.linkMap.Get(path)
	if exists { //TODO: Move this check inside the LinkMap, return delta from update fn
		oldEntry = &models.Entry{
			Path:   path,
			Target: oldTarget,
		}

		if err := h.linkMap.Update(newEntry.Path, targetUrl); err != nil {
			middleware.SetError(r, http.StatusInternalServerError, fmt.Errorf("error updating link %s: %w", newEntry.Path, err))
			return
		}
	} else {
		if err := h.linkMap.Put(path, targetUrl); err != nil {
			middleware.SetError(r, http.StatusInternalServerError, fmt.Errorf("error adding link %s: %w", newEntry.Path, err))
			return
		}
	}

	update := models.UpdateDelta{
		Old: oldEntry,
		New: newEntry,
	}

	utils.RespondJSON(w, r, http.StatusOK, update)
}

func (h *ApiHandler) deleteLink(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "path")
	if err := h.linkMap.Delete(path); err != nil {
		middleware.SetError(r, http.StatusInternalServerError, fmt.Errorf("error deleting link %s: %w", path, err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ApiHandler) getAll(w http.ResponseWriter, r *http.Request) {
	allLinks := h.linkMap.GetAll()
	utils.RespondJSON(w, r, http.StatusOK, allLinks)
}

func (h *ApiHandler) getAllForAlfred(w http.ResponseWriter, r *http.Request) {
	alfredResponse := buildAlfredResponse(h.linkMap.GetAll())
	utils.RespondJSON(w, r, http.StatusOK, alfredResponse)
}

func (h *ApiHandler) getLink(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "path")
	target, exists := h.linkMap.Get(path)
	if !exists {
		middleware.SetError(r, http.StatusNotFound, fmt.Errorf("path %s has no target", path))
		return
	}

	utils.RespondJSON(w, r, http.StatusOK, models.Entry{
		Path:   path,
		Target: target,
	})
}

func (h *ApiHandler) search(w http.ResponseWriter, r *http.Request) {
	options := h.linkMap.GetAllKeys()
	query := r.URL.Query().Get("query")
	isAlfredRequest := r.URL.Query().Get("isAlfred") == "true"
	hits := search.StringSearch(query, options)
	keyHits := make([]string, len(hits))
	for i, hit := range hits {
		keyHits[i] = hit.Value
	}

	hitMap := h.linkMap.GetFiltered(keyHits)

	if isAlfredRequest {
		resp := buildAlfredResponse(hitMap)
		utils.RespondJSON(w, r, http.StatusOK, resp)
		return
	}

	utils.RespondJSON(w, r, http.StatusOK, hitMap)
}

func buildAlfredResponse(mapItems map[string]string) *alfredResponse {
	items := make([]alfredItem, len(mapItems))
	for key, val := range mapItems {
		item := alfredItem{
			Uid:          key,
			Title:        key,
			Subtitle:     val,
			Arg:          val,
			Autocomplete: key,
		}
		items = append(items, item)
	}

	return &alfredResponse{
		Items: items,
	}
}
