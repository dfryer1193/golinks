package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	var port int
	var configDir string
	var configFile string
	configDir := os.Getenv("XDG_CONFIG_HOME")
	port := *flag.Int("port", 8080, "The port to listen on")
	configFile := flag.String("config", configDir+"links", "Location of the config file")
	fmt.Printf("Starting http server:\nListening on port :%d\nUsing config file: %s", port, *configFile)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}
