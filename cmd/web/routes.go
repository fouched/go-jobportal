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
	mux.Get("/login", app.Login)
	mux.Post("/login", app.LoginPost)
	mux.Get("/logout", app.Logout)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(app.Auth)
		mux.Get("/dashboard", app.Dashboard)
	})

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
