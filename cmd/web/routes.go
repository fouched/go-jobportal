package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(SessionLoad)
	mux.Use(middleware.Recoverer)

	mux.Get("/", app.Home)
	mux.Get("/register", app.ShowRegister)
	mux.Post("/register/new", app.RegisterNew)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
