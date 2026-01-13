package handler

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"

	"github.com/dfryer1193/mjolnir/utils/errorx"
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

type FrontendHandler struct{}

func NewFrontendHandler() *FrontendHandler {
	return &FrontendHandler{}
}

func (h *FrontendHandler) serveHomepage(w http.ResponseWriter, r *http.Request) *errorx.ApiError {
	file, err := serveEmbeddedContent(w, r, INDEX)
	if err != nil {
		return errorx.InternalServerErr(err)
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return errorx.InternalServerErr(err)
	}

	sizeStr := fmt.Sprintf("%d", fileStat.Size())

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Length", sizeStr)
	w.Header().Set("Content-Type", "text/html")
	http.ServeContent(w, r, fileStat.Name(), fileStat.ModTime(), file.(io.ReadSeeker))

	return nil
}

func (h *FrontendHandler) serveFavicon(w http.ResponseWriter, r *http.Request) *errorx.ApiError {
	file, err := serveEmbeddedContent(w, r, FAVICON)
	if err != nil {
		return errorx.InternalServerErr(err)
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return errorx.InternalServerErr(err)
	}

	sizeStr := fmt.Sprintf("%d", fileStat.Size())

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Content-Length", sizeStr)
	http.ServeContent(w, r, fileStat.Name(), fileStat.ModTime(), file.(io.ReadSeeker))

	return nil
}

func (h *FrontendHandler) serveStyles(w http.ResponseWriter, r *http.Request) *errorx.ApiError {
	file, err := serveEmbeddedContent(w, r, STYLES)
	if err != nil {
		return errorx.InternalServerErr(err)
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return errorx.InternalServerErr(err)
	}

	sizeStr := fmt.Sprintf("%d", fileStat.Size())

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Length", sizeStr)
	w.Header().Set("Content-Type", "text/css")
	http.ServeContent(w, r, fileStat.Name(), fileStat.ModTime(), file.(io.ReadSeeker))

	return nil
}

func (h *FrontendHandler) serveNewForm(w http.ResponseWriter, r *http.Request) *errorx.ApiError {
	file, err := serveEmbeddedContent(w, r, NEW)
	if err != nil {
		return errorx.InternalServerErr(err)
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return errorx.InternalServerErr(err)
	}

	sizeStr := fmt.Sprintf("%d", fileStat.Size())

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Length", sizeStr)
	w.Header().Set("Content-Type", "text/html")
	http.ServeContent(w, r, fileStat.Name(), fileStat.ModTime(), file.(io.ReadSeeker))

	return nil
}

// serveEmbeddedContent serves static content embedded in the binary. The returned file **MUST** be closed by the caller.
func serveEmbeddedContent(w http.ResponseWriter, r *http.Request, contentKey ContentName) (fs.File, error) {
	filename := staticContent[contentKey]
	file, err := content.Open(filename)
	if err != nil {
		return nil, err
	}

	return file, nil
}
