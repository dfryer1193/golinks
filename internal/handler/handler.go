package handler

import (
	"embed"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/dfryer1193/golinks/internal/links"
)

//go:embed static/*
var content embed.FS

const embedDir = "static/"
const list = embedDir + "index.html"
const css = embedDir + "styles.css"
const newLink = embedDir + "new.html"

// GolinkHandler handles all incoming/outgoing http requests for go links.
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

// NewHandler returns a reference to a new instance of a GolinkHandler
func NewHandler(linkMapPtr *links.LinkMap) *GolinkHandler {
	return &GolinkHandler{
		linkMap: linkMapPtr,
	}
}

// ServeHTTP handles the base logic of handling http requests for the GolinkHandler
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

	staticPathHandler := h.getStaticPathMap()[path]

	if h.getStaticPathMap()[path] != nil {
		staticPathHandler(w, req)
		return
	}

	target, exists := h.linkMap.Get(path)

	if exists {
		http.Redirect(w, req, target.String(), http.StatusTemporaryRedirect)
		return
	}

	h.serveNewForm(w, req)
}

func (h *GolinkHandler) serveList(w http.ResponseWriter, req *http.Request) {
	serveEmbeddedHtml(list, w)
}

func (h *GolinkHandler) serveStyles(w http.ResponseWriter, req *http.Request) {
	cssBytes, err := content.ReadFile(css)
	if err != nil {
		http.Error(w, "Error reading styles.css", http.StatusInternalServerError)
		slog.Error("Error reading embedded styles.css", "error", err)
		return
	}

	w.Header().Set("Content-Type", "text/css")

	_, err = w.Write(cssBytes)
	if err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		slog.Error("Error writing css response", "error", err)
	}
}

func (h *GolinkHandler) serveNewForm(w http.ResponseWriter, req *http.Request) {
	serveEmbeddedHtml(newLink, w)
}

func (h *GolinkHandler) getAllRedirects(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(h.linkMap.GetAllAsString())
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		slog.Error("Error encoding link map to JSON", "error", err)
		return
	}
}

func (h *GolinkHandler) getStaticPathMap() map[string]func(http.ResponseWriter, *http.Request) {
	return map[string]func(w http.ResponseWriter, req *http.Request){
		"":            h.serveList,
		"favicon.ico": func(w http.ResponseWriter, req *http.Request) {}, // ignore favicon requests
		"redirects":   h.getAllRedirects,
		"styles.css":  h.serveStyles,
		"update":      h.serveNewForm,
	}
}

func (h *GolinkHandler) handlePost(w http.ResponseWriter, req *http.Request) {
	var oldPathAndTarget *pathAndTarget
	path, target, err := extractPathAndTarget(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	if _, exists := h.getStaticPathMap()[path]; exists {
		w.WriteHeader(http.StatusBadRequest)
		return
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

func serveEmbeddedHtml(filename string, w http.ResponseWriter) {
	htmlBytes, err := content.ReadFile(filename)
	if err != nil {
		http.Error(w, "Error reading index.html", http.StatusInternalServerError)
		slog.Error("Error reading embedded index.html", "error", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	_, err = w.Write(htmlBytes)
	if err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		slog.Error("Error writing html response", "error", err)
	}
}
