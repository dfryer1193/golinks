package handler

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"

	"github.com/dfryer1193/golinks/internal/links"
	"github.com/dfryer1193/golinks/internal/search"
	"github.com/dfryer1193/golinks/models"
	"github.com/dfryer1193/mjolnir/utils/errorx"
	"github.com/dfryer1193/mjolnir/utils/httpx"
	"github.com/go-chi/chi/v5"
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
	linkMap links.LinkMap
}

func NewApiHandler(linkMap links.LinkMap) *ApiHandler {
	return &ApiHandler{linkMap: linkMap}
}

func (h *ApiHandler) postLink(w http.ResponseWriter, r *http.Request) *errorx.ApiError {
	path := chi.URLParam(r, "path")
	target := &struct {
		Target string `json:"target"`
	}{}
	_, err := httpx.DecodeJSON(r, target)
	if err != nil {
		return errorx.BadRequestErr(fmt.Errorf("invalid request body: %w", err))
	}
	
	// Validate target is not empty
	if target.Target == "" {
		return errorx.BadRequestErr(fmt.Errorf("target is required"))
	}
	
	newEntry := &models.Entry{
		Path:   path,
		Target: target.Target,
	}

	// Parse and validate URL
	targetUrl, err := url.Parse(target.Target)
	if err != nil {
		return errorx.BadRequestErr(fmt.Errorf("invalid target URL %s: %w", target.Target, err))
	}
	
	// Require absolute URLs (must have scheme and host)
	if targetUrl.Scheme == "" || targetUrl.Host == "" {
		return errorx.BadRequestErr(fmt.Errorf("target must be an absolute URL with scheme and host (e.g., https://example.com)"))
	}

	var oldEntry *models.Entry
	oldTarget, exists := h.linkMap.Get(path)
	if exists { //TODO: Move this check inside the LinkMap, return delta from update fn
		oldEntry = &models.Entry{
			Path:   path,
			Target: oldTarget,
		}

		if err := h.linkMap.Update(newEntry.Path, targetUrl); err != nil {
			return errorx.InternalServerErr(fmt.Errorf("error updating link %s: %w", newEntry.Path, err))
		}
	} else {
		if err := h.linkMap.Put(path, targetUrl); err != nil {
			return errorx.InternalServerErr(fmt.Errorf("error creating link %s: %w", newEntry.Path, err))
		}
	}

	update := models.UpdateDelta{
		Old: oldEntry,
		New: newEntry,
	}

	httpx.RespondJSON(w, r, http.StatusOK, update)

	return nil
}

func (h *ApiHandler) deleteLink(w http.ResponseWriter, r *http.Request) *errorx.ApiError {
	path := chi.URLParam(r, "path")
	if err := h.linkMap.Delete(path); err != nil {
		return errorx.InternalServerErr(fmt.Errorf("error deleting link %s: %w", path, err))
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (h *ApiHandler) getAll(w http.ResponseWriter, r *http.Request) *errorx.ApiError {
	allLinks := h.linkMap.GetAll()
	httpx.RespondJSON(w, r, http.StatusOK, allLinks)

	return nil
}

func (h *ApiHandler) getAllForAlfred(w http.ResponseWriter, r *http.Request) *errorx.ApiError {
	alfredResponse := buildAlfredResponse(h.linkMap.GetAll())
	httpx.RespondJSON(w, r, http.StatusOK, alfredResponse)

	return nil
}

func (h *ApiHandler) getLink(w http.ResponseWriter, r *http.Request) *errorx.ApiError {
	path := chi.URLParam(r, "path")
	target, exists := h.linkMap.Get(path)
	if !exists {
		return errorx.BadRequestErr(fmt.Errorf("link not found for path: %s", path))
	}

	httpx.RespondJSON(w, r, http.StatusOK, models.Entry{
		Path:   path,
		Target: target,
	})

	return nil
}

func (h *ApiHandler) search(w http.ResponseWriter, r *http.Request) *errorx.ApiError {
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
		httpx.RespondJSON(w, r, http.StatusOK, resp)
		return nil
	}

	httpx.RespondJSON(w, r, http.StatusOK, hitMap)

	return nil
}

func (h *ApiHandler) exportLinks(w http.ResponseWriter, r *http.Request) *errorx.ApiError {
	allLinks := h.linkMap.GetAll()

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=links")

	for path, target := range allLinks {
		_, err := fmt.Fprintf(w, "%s %s\n", path, target)
		if err != nil {
			return errorx.InternalServerErr(fmt.Errorf("error writing link %s to response: %w", path, err))
		}
	}

	return nil
}

func (h *ApiHandler) importLinks(w http.ResponseWriter, r *http.Request) *errorx.ApiError {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return errorx.BadRequestErr(fmt.Errorf("invalid content type: %w", err))
	}

	if mediaType != "text/plain" {
		return errorx.BadRequestErr(fmt.Errorf("unsupported content type: %s; expected text/plain", mediaType))
	}

	err = h.linkMap.ReplaceAll(r.Body)
	if err != nil {
		return errorx.InternalServerErr(fmt.Errorf("error importing links: %w", err))
	}
	w.WriteHeader(http.StatusNoContent)

	return nil
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
