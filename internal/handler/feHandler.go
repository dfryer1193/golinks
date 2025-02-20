package handler

import (
	"embed"
	"fmt"
	"github.com/dfryer1193/mjolnir/middleware"
	"io"
	"net/http"
)

//go:embed static/*
var content embed.FS

type ContentName int

const (
	INDEX ContentName = iota
	NEW
	STYLES
	FAVICON
)

var staticContent = map[ContentName]string{
	INDEX:   "static/index.html",
	NEW:     "static/new.html",
	STYLES:  "static/styles.css",
	FAVICON: "static/favicon.ico",
}

type FeHandler struct{}

func NewFeHandler() *FeHandler {
	return &FeHandler{}
}

func (h *FeHandler) serveHomepage(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedContent(w, r, INDEX)
}

func (h *FeHandler) serveFavicon(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedContent(w, r, FAVICON)
}

func (h *FeHandler) serveStyles(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedContent(w, r, STYLES)
}

func (h *FeHandler) serveNewForm(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedContent(w, r, NEW)
}

func serveEmbeddedContent(w http.ResponseWriter, r *http.Request, contentKey ContentName) {
	filename := staticContent[contentKey]
	file, err := content.Open(filename)
	if err != nil {
		middleware.SetInternalError(r, fmt.Errorf("error opening embedded file %s: %w", filename, err))
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		middleware.SetInternalError(r, fmt.Errorf("cannot get file info for embedded file %s: %w", filename, err))
	}

	contentType := ""
	switch {
	case stat.Name() == "favicon.ico":
		contentType = "image/x-icon"
	case stat.Name() == "styles.css":
		contentType = "text/css"
	case stat.Name() == "new.html" || stat.Name() == "index.html":
		contentType = "text/html"
	}

	w.Header().Set("Content-Type", contentType)
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file.(io.ReadSeeker))
}
