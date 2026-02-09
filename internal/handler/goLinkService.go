package handler

import (
	"net/http"

	"github.com/dfryer1193/golinks/config"
	"github.com/dfryer1193/golinks/internal/links"
	"github.com/dfryer1193/golinks/internal/links/storage"
	"github.com/dfryer1193/mjolnir/utils/errorx"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// GolinkHandler handles all incoming/outgoing http requests for go links.
type GolinkHandler struct {
	linkMap         links.LinkMap
	apiHandler      *ApiHandler
	frontendHandler *FrontendHandler
}

// NewGoLinkService returns a reference to a new instance of a GolinkHandler
func NewGoLinkService(router *chi.Mux, cfg *config.Config) {
	var linkMap links.LinkMap

	if cfg.StorageType == storage.SQLITE {
		linkMap = links.NewBaseLinkMap(cfg.StorageType, cfg.ConfigFile)
	} else {
		linkMap = links.NewCachingLinkMap(cfg.StorageType, cfg.ConfigFile)
	}

	apiHandler := NewApiHandler(linkMap)
	frontendHandler := NewFrontendHandler()
	service := &GolinkHandler{
		linkMap:         linkMap,
		apiHandler:      apiHandler,
		frontendHandler: frontendHandler,
	}

	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/all", errorx.ErrorHandler(apiHandler.getAll))
		r.Get("/all/alfred", errorx.ErrorHandler(apiHandler.getAllForAlfred))
		r.Get("/search", errorx.ErrorHandler(apiHandler.search))
		r.Get("/links/{path}", errorx.ErrorHandler(apiHandler.getLink))
		r.Post("/links/{path}", errorx.ErrorHandler(apiHandler.postLink))
		r.Delete("/links/{path}", errorx.ErrorHandler(apiHandler.deleteLink))
		r.Get("/export", errorx.ErrorHandler(apiHandler.exportLinks))
		r.Post("/import", errorx.ErrorHandler(apiHandler.importLinks))
	})

	router.Route("/", func(r chi.Router) {
		r.Use(noCacheMiddleware)
		r.Get("/", errorx.ErrorHandler(frontendHandler.serveHomepage))
		r.Get("/favicon.ico", errorx.ErrorHandler(frontendHandler.serveFavicon))
		r.Get("/styles.css", errorx.ErrorHandler(frontendHandler.serveStyles))
		r.Get("/update", errorx.ErrorHandler(frontendHandler.serveNewForm))
		r.Get("/*", errorx.ErrorHandler(service.handleGet))
	})
}

func (h *GolinkHandler) handleGet(w http.ResponseWriter, r *http.Request) *errorx.ApiError {
	path := chi.URLParam(r, "*")

	target, exists := h.linkMap.Get(path)

	if exists {
		log.Debug().Str("target", target).Msg("Shortcut found! Redirecting...")
		http.Redirect(w, r, target, http.StatusTemporaryRedirect)
		return nil
	}

	return h.frontendHandler.serveNewForm(w, r)
}
