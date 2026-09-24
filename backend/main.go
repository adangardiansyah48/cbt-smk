package main

import (
	"log"
	"os"
	"strings"

	cbtapp "cbt-api/app"
)

func main() {
	app := cbtapp.BuildApp()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	log.Printf("listening %s", port)
	log.Fatal(app.Listen(port))
}
