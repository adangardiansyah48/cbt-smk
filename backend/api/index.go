package handler

import (
	"net/http"

	cbtapp "cbt-api/internal/app"

	"github.com/gofiber/adaptor/v2"
)

var h http.Handler

func init() {
	app := cbtapp.BuildApp()
	h = adaptor.FiberApp(app)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	h.ServeHTTP(w, r)
}
