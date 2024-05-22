package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type GolinkHandler struct {
	linkMap map[string]url.URL
}

func NewHandler(linkMapPtr map[string]url.URL) *GolinkHandler {
	return &GolinkHandler{
		linkMap: linkMapPtr,
	}
}

func (h GolinkHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		fmt.Println("Cannot serve non-get requests")
		return
	}

	path := strings.TrimPrefix(req.URL.Path, "/")
	target, exists := h.linkMap[path]

	fmt.Printf("Handling request. path={%s} target={%s} target-exists={%t}\n",
		path,
		target.String(),
		exists,
	)

	if exists {
		http.Redirect(w, req, target.String(), http.StatusTemporaryRedirect)
		return
	}

	http.NotFound(w, req)
}
