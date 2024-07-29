package handler

import (
	"encoding/json"
	"github.com/dfryer1193/golinks/internal/search"
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
)

type linkTarget struct {
	Target string `json:"target"`
}

type pathAndTarget struct {
	Path   string `json:"path"`
	Target string `json:"target"`
}

type targetUpdate struct {
	Old *pathAndTarget `json:"old"`
	New *pathAndTarget `json:"new"`
}

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

const apiPath = "/api/v1/"

func (h *GolinkHandler) handleV1ApiRequest(w http.ResponseWriter, r *http.Request) {
	strippedPath := strings.TrimPrefix(r.URL.Path, apiPath)
	log.Debug().Str("path", r.URL.Path).Str("strippedPath", strippedPath).Msg("Handling ")
	switch r.Method {
	case http.MethodGet:
		h.handleApiGet(w, r, strippedPath)
	case http.MethodPost:
		h.handleApiPost(w, r)
	case http.MethodDelete:
		h.handleApiDelete(w, strippedPath)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *GolinkHandler) handleApiGet(w http.ResponseWriter, r *http.Request, strippedPath string) {
	if strippedPath == "all" {
		log.Debug().Msg("Getting all entries...")
		h.getAll(w)
		return
	} else if strippedPath == "all/alfred" {
		log.Debug().Msg("Getting all entries with alfred item formatting...")
		h.getAllForAlfred(w)
		return
	} else if strippedPath == "search" {
		log.Debug().Msg("Getting search results...")
		h.search(w, r)
		return
	}

	h.get(w, strippedPath)
}

func (h *GolinkHandler) handleApiPost(w http.ResponseWriter, r *http.Request) {
	var oldPathAndTarget *pathAndTarget
	path, target, err := extractPathAndTarget(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if target == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	oldTarget, exists := h.linkMap.Get(path)
	if exists { //TODO: Move this check inside the LinkMap, return delta from update fn
		oldPathAndTarget = &pathAndTarget{
			Path:   path,
			Target: oldTarget,
		}

		err := h.linkMap.Update(path, target)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatal().Err(err).Msg("Encountered an error writing config")
		}
	} else {
		err = h.linkMap.Put(path, target)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatal().Err(err).Msg("Encountered an error writing config")
		}
	}

	update := targetUpdate{
		Old: oldPathAndTarget,
		New: &pathAndTarget{
			Path:   path,
			Target: target.String(),
		},
	}

	responseBytes, err := json.Marshal(update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	_, err = w.Write(responseBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *GolinkHandler) handleApiDelete(w http.ResponseWriter, strippedPath string) {
	err := h.linkMap.Delete(strippedPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GolinkHandler) getAll(w http.ResponseWriter) {
	err := json.NewEncoder(w).Encode(h.linkMap.GetAll())
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		log.Err(err).Msg("Error encoding link map to JSON")
		return
	}
}

func (h *GolinkHandler) getAllForAlfred(w http.ResponseWriter) {
	sendAlfredResponse(w, h.linkMap.GetAll())
}

func (h *GolinkHandler) get(w http.ResponseWriter, strippedPath string) {
	target, exists := h.linkMap.Get(strippedPath)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
	}

	jsonBytes, err := json.Marshal(target)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Err(err).Msg("Error encoding JSON")
		return
	}

	_, err = w.Write(jsonBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Err(err).Msg("Error writing response body")
		return
	}
}

func (h *GolinkHandler) search(w http.ResponseWriter, r *http.Request) {
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
		sendAlfredResponse(w, hitMap)
	}

	err := json.NewEncoder(w).Encode(hitMap)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		log.Err(err).Msg("Error encoding link map to JSON")
		return
	}
}

func extractPathAndTarget(req *http.Request) (string, *url.URL, error) {
	path := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, apiPath), "/") // Don't allow trailing slashes for shortcuts
	target, err := getBody(req)
	if err != nil {
		return "", nil, err
	}

	targetURL, err := url.Parse(target.Target)
	if err != nil {
		return "", nil, err
	}

	return path, targetURL, nil
}

func getBody(req *http.Request) (*linkTarget, error) {
	var body linkTarget
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}

func sendAlfredResponse(w http.ResponseWriter, mapItems map[string]string) {
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

	r := alfredResponse{
		Items: items,
	}

	jsonBytes, err := json.Marshal(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Err(err).Msg("Error encoding JSON")
		return
	}

	_, err = w.Write(jsonBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Err(err).Msg("Error writing response body")
		return
	}
}
