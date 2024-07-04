package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/dfryer1193/golinks/internal/links"
)

// GolinkHandler handles all incoming/outgoing http requests for go links.
type GolinkHandler struct {
	linkMap *links.LinkMap
}

// NewGolinkHandler returns a reference to a new instance of a GolinkHandler
func NewGolinkHandler(linkMapPtr *links.LinkMap) *GolinkHandler {
	return &GolinkHandler{
		linkMap: linkMapPtr,
	}
}

// ServeHTTP handles the base logic of handling http requests for the GolinkHandler
func (h *GolinkHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	slog.Info(req.Method + " " + req.URL.Path)

	if strings.HasPrefix(req.URL.Path, apiPath) {
		h.handleV1ApiRequest(w, req)
		return
	}

	if feHandlerFunc, exists := h.getServedFePaths()[req.URL.Path]; exists {
		feHandlerFunc(w)
		return
	}

	switch req.Method {
	case http.MethodGet:
		h.handleGet(w, req)
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

	h.serveNewForm(w)
}
