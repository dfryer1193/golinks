package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/dfryer1193/golinks/internal/handler"
	"github.com/dfryer1193/golinks/internal/links"
)

func main() {
	var port int
	var configFile string
	flag.IntVar(&port, "port", 8080, "The port to listen on")
	flag.StringVar(&configFile, "config", "", "Location of the config file")

	fmt.Printf("Starting http server:\nListening on port :%d\n", port)

	redirector := handler.NewHandler(links.NewLinkMap(configFile))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), *redirector); err != nil {
		log.Fatal(err)
	}
}
