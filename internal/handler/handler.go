package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dfryer1193/golinks/internal/links"
)

type GolinkHandler struct {
	linkMap *links.LinkMap
}

func NewHandler(linkMapPtr *links.LinkMap) *GolinkHandler {
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
	target, exists := h.linkMap.Get(path)

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
