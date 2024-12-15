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
		mux.Get("/recruiter-profile", app.RecruiterProfile)
		mux.Post("/recruiter-profile/add", app.RecruiterProfileUpdate)
		mux.Get("/dashboard/add", app.JobPostNew)
		mux.Post("/dashboard/add-save", app.JobPostSave)
	})

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	uploadsServer := http.FileServer(http.Dir("./uploads/"))
	mux.Handle("/uploads/*", http.StripPrefix("/uploads", uploadsServer))

	return mux
}
