package api

import (
	"net/http"

	"github.com/go-chi/chi"
)

// router is the main api router
var router = chi.NewRouter()

// Router returns the api router
func Router() http.Handler {
	router.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})

	registerRoutes()

	return router
}

func registerRoutes() {
	router.Route("/v1/", func(r chi.Router) {
		r.Mount("/app", appsV1())
	})
}

func appsV1() http.Handler {
	h := chi.NewRouter()
	h.Group(func(r chi.Router) {
		r.Get("/{id}", getUser)
		r.Post("/{id}", updateUser)
	})

	return h
}
