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
	mux.Post("/register/save", app.RegisterSave)
	mux.Get("/login", app.Login)
	mux.Post("/login", app.LoginPost)
	mux.Get("/logout", app.Logout)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(app.Auth)

		mux.Get("/dashboard", app.Dashboard)
		mux.Post("/dashboard", app.Dashboard)

		mux.Get("/job-post", app.JobPost)
		mux.Post("/job-post/edit/{id}", app.JobPostEdit)
		mux.Post("/job-post/save", app.JobPostSave)

		mux.Get("/recruiter-profile", app.RecruiterProfile)
		mux.Post("/recruiter-profile/save", app.RecruiterProfileSave)

		mux.Get("/job-seeker-profile", app.JobSeekerProfile)
		mux.Get("/job-seeker-profile/{id}", app.JobSeekerProfile)
		mux.Post("/job-seeker-profile/save", app.JobSeekerProfileSave)

		mux.Get("/job-details/{id}", app.JobDetails)
		mux.Post("/job-details/apply/{id}", app.JobDetailsApply)
		mux.Post("/job-details/save/{id}", app.JobDetailsSave)

		mux.Get("/saved-jobs", app.SavedJobs)
		mux.Get("/download-resume/{id}", app.DownloadResume)
	})

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	uploadsServer := http.FileServer(http.Dir("./uploads/"))
	mux.Handle("/uploads/*", http.StripPrefix("/uploads", uploadsServer))

	return mux
}
