package main

import "net/http"

// Home displays the home page
func (app *application) Home(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "home", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// Home displays the registration page
func (app *application) Register(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "register", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}
