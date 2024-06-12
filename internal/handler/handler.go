package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/dfryer1193/golinks/internal/links"
)

type GolinkHandler struct {
	linkMap *links.LinkMap
}

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

func NewHandler(linkMapPtr *links.LinkMap) *GolinkHandler {
	return &GolinkHandler{
		linkMap: linkMapPtr,
	}
}

func (h *GolinkHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	slog.Info(req.Method + " " + req.URL.Path)
	switch req.Method {
	case http.MethodGet:
		h.handleGet(w, req)
	case http.MethodPost:
		h.handlePost(w, req)
	case http.MethodDelete:
		h.handleDelete(w, req)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *GolinkHandler) handleGet(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/")
	target, exists := h.linkMap.Get(path)

	if exists {
		http.Redirect(w, req, target.String(), http.StatusTemporaryRedirect)
		return
	}

	http.NotFound(w, req)
}

func (h *GolinkHandler) handlePost(w http.ResponseWriter, req *http.Request) {
	var oldPathAndTarget *pathAndTarget
	path, target, err := extractPathAndTarget(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	oldTarget, exists := h.linkMap.Get(path)
	if exists {
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
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(responseBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *GolinkHandler) handleDelete(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/")
	err := h.linkMap.Delete(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func extractPathAndTarget(req *http.Request) (string, *url.URL, error) {
	path := strings.TrimPrefix(req.URL.Path, "/")
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
