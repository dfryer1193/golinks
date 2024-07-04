package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
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

const apiPath = "/api/v1/"

func (h *GolinkHandler) handleV1ApiRequest(w http.ResponseWriter, r *http.Request) {
	strippedPath := strings.TrimPrefix(r.URL.Path, apiPath)
	switch r.Method {
	case http.MethodGet:
		h.handleApiGet(w, strippedPath)
	case http.MethodPost:
		h.handleApiPost(w, r)
	case http.MethodDelete:
		h.handleApiDelete(w, strippedPath)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *GolinkHandler) handleApiGet(w http.ResponseWriter, strippedPath string) {
	if strippedPath == "all" {
		h.getAll(w)
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
			Target: oldTarget.String(),
		}

		err := h.linkMap.Update(path, target)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("Encountered an error writing config", "error", err)
			panic(1)
		}
	} else {
		err = h.linkMap.Put(path, target)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("Encountered an error writing config", "error", err)
			panic(1)
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
	err := json.NewEncoder(w).Encode(h.linkMap.GetAllAsString())
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		slog.Error("Error encoding link map to JSON", "error", err)
		return
	}
}

func (h *GolinkHandler) get(w http.ResponseWriter, strippedPath string) {
	url, exists := h.linkMap.Get(strippedPath)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
	}

	jsonBytes, err := json.Marshal(url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Error encoding JSON", "error", err)
		return
	}

	_, err = w.Write(jsonBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Error writing response body", "error", err)
		return
	}
}

func extractPathAndTarget(req *http.Request) (string, *url.URL, error) {
	path := strings.TrimPrefix(req.URL.Path, apiPath)
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
