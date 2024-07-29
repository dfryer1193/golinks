package handler

import (
	"embed"
	"net/http"

	"github.com/rs/zerolog/log"
)

//go:embed static/*
var content embed.FS

const embedDir = "static/"
const list = embedDir + "index.html"
const css = embedDir + "styles.css"
const newLink = embedDir + "new.html"
const favicon = embedDir + "favicon.ico"

func (h *GolinkHandler) getServedFePaths() map[string]func(w http.ResponseWriter) {
	return map[string]func(w http.ResponseWriter){
		"/":                  h.serveList,
		"/favicon.ico":       h.serveFavicon,
		"/static/styles.css": h.serveStyles,
		"/static/update":     h.serveNewForm,
		"/static/index.html": h.serveList,
		"/static/new.html":   h.serveNewForm,
	}
}

func (h *GolinkHandler) serveList(w http.ResponseWriter) {
	serveEmbeddedHtml(list, w)
}

func (h *GolinkHandler) serveStyles(w http.ResponseWriter) {
	cssBytes, err := content.ReadFile(css)
	if err != nil {
		http.Error(w, "Error reading styles.css", http.StatusInternalServerError)
		log.Err(err).Msg("Error reading embedded styles.css")
		return
	}

	w.Header().Set("Content-Type", "text/css")

	_, err = w.Write(cssBytes)
	if err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		log.Err(err).Msg("Error writing css response")
	}
}

func (h *GolinkHandler) serveNewForm(w http.ResponseWriter) {
	serveEmbeddedHtml(newLink, w)
}

func (h *GolinkHandler) serveFavicon(w http.ResponseWriter) {
	iconBytes, err := content.ReadFile(favicon)
	if err != nil {
		http.Error(w, "Error serving favicon", http.StatusInternalServerError)
		log.Err(err).Msg("Error reading embedded favicon")
	}

	_, err = w.Write(iconBytes)
	if err != nil {
		http.Error(w, "Error writing favicon", http.StatusInternalServerError)
		log.Err(err).Msg("Error writing favicon to response")
	}
}

func serveEmbeddedHtml(filename string, w http.ResponseWriter) {
	htmlBytes, err := content.ReadFile(filename)
	if err != nil {
		http.Error(w, "Error reading index.html", http.StatusInternalServerError)
		log.Err(err).Msg("Error reading embedded index.html")
		return
	}

	_, err = w.Write(htmlBytes)
	if err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		log.Err(err).Msg("Error writing html response")
	}
}
