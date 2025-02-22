package handler

import (
	"github.com/dfryer1193/golinks/config"
	"github.com/dfryer1193/golinks/internal/links"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

// GolinkHandler handles all incoming/outgoing http requests for go links.
type GolinkHandler struct {
	linkMap         *links.LinkMap
	apiHandler      *ApiHandler
	frontendHandler *FrontendHandler
}

// NewGoLinkService returns a reference to a new instance of a GolinkHandler
func NewGoLinkService(router *chi.Mux, cfg *config.Config) {
	linkMap := links.NewLinkMap(cfg.StorageType, cfg.ConfigFile)
	apiHandler := NewApiHandler(linkMap)
	frontendHandler := NewFrontendHandler()
	service := &GolinkHandler{
		linkMap:         linkMap,
		apiHandler:      apiHandler,
		frontendHandler: frontendHandler,
	}

	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/all", apiHandler.getAll)
		r.Get("/all/alfred", apiHandler.getAllForAlfred)
		r.Get("/search", apiHandler.search)
		r.Get("/links/{path}", apiHandler.getLink)
		r.Post("/links/{path}", apiHandler.postLink)
		r.Delete("/links/{path}", apiHandler.deleteLink)
	})

	router.Route("/", func(r chi.Router) {
		r.Use(noCacheMiddleware)
		r.Get("/", frontendHandler.serveHomepage)
		r.Get("/favicon.ico", frontendHandler.serveFavicon)
		r.Get("/styles.css", frontendHandler.serveStyles)
		r.Get("/update", frontendHandler.serveNewForm)
		r.Get("/{path}", service.handleGet)
	})
}

func (h *GolinkHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "path")

	target, exists := h.linkMap.Get(path)

	if exists {
		log.Debug().Str("target", target).Msg("Shortcut found! Redirecting...")
		http.Redirect(w, r, target, http.StatusTemporaryRedirect)
		return
	}

	h.frontendHandler.serveNewForm(w, r)
}
