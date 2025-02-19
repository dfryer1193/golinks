package handler

import (
	"embed"
	"fmt"
	"github.com/dfryer1193/mjolnir/middleware"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

//go:embed static/*
var content embed.FS

const embedDir = "./internal/handler/static/"

type FeHandler struct {
	EmbeddedRoot string
}

func NewFeHandler() *FeHandler {
	root, _ := filepath.Abs(embedDir)
	log.Debug().Str("root", root).Msg("Embedded root")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		log.Fatal().Msg("Failed to find embedded static directory")
	}

	return &FeHandler{
		EmbeddedRoot: root,
	}
}

func (h *FeHandler) serveHomepage(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(h.EmbeddedRoot + "/index.html"); os.IsNotExist(err) {
		middleware.SetError(r, http.StatusInternalServerError, fmt.Errorf("embedded index.html not found"))
		return
	}
	http.ServeFile(w, r, h.EmbeddedRoot+"/index.html")
}

func (h *FeHandler) serveFavicon(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(h.EmbeddedRoot + "/favicon.ico"); os.IsNotExist(err) {
		middleware.SetError(r, http.StatusInternalServerError, fmt.Errorf("embedded favicon not found"))
		return
	}
	http.ServeFile(w, r, h.EmbeddedRoot+"/favicon.ico")
}

func (h *FeHandler) serveStyles(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(h.EmbeddedRoot + "/styles.css"); os.IsNotExist(err) {
		middleware.SetError(r, http.StatusInternalServerError, fmt.Errorf("embedded styles.css not found"))
		return
	}
	http.ServeFile(w, r, h.EmbeddedRoot+"/styles.css")
}

func (h *FeHandler) serveNewForm(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(h.EmbeddedRoot + "/new.html"); os.IsNotExist(err) {
		middleware.SetError(r, http.StatusInternalServerError, fmt.Errorf("embedded new.html not found"))
		return
	}

	http.ServeFile(w, r, h.EmbeddedRoot+"/new.html")
}
