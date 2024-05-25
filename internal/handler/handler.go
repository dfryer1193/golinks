package handler

import (
	"encoding/json"
	"fmt"
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

func NewHandler(linkMapPtr *links.LinkMap) *GolinkHandler {
	return &GolinkHandler{
		linkMap: linkMapPtr,
	}
}

func (h *GolinkHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("Handling request %s %s\n", req.Method, req.URL.Path)
	switch req.Method {
	case http.MethodGet:
		h.handleGet(w, req)
	case http.MethodPost:
		h.handlePost(w, req)
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
	path := strings.TrimPrefix(req.URL.Path, "/")
	_, exists := h.linkMap.Get(path)

	if exists {
		fmt.Println("Path exists, discarding update.")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	target, err := getBody(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("Received update: key:{%s} target:{%s}\n", path, target)

	targetURL, err := url.Parse(target.Target)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.linkMap.Put(path, *targetURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Encountered an error writing updates to file: %s\n", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getBody(req *http.Request) (*linkTarget, error) {
	var body linkTarget
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}
