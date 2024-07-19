package handler

import (
	"net/http"
	"strings"

	"github.com/dfryer1193/golinks/internal/links"
	"github.com/rs/zerolog/log"
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
	log.Info().Msg(req.Method + " " + req.URL.Path)
	log.Debug().Msg(req.Method + " " + req.Host)

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
	path := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/") // Disallow trailing slashes

	target, exists := h.linkMap.Get(path)

	if exists {
		log.Debug().Str("target", target).Msg("Shortcut found! Redirecting...")
		http.Redirect(w, req, target, http.StatusTemporaryRedirect)
		return
	}

	h.serveNewForm(w)
}
