package main

import (
	"github.com/fouched/go-jobportal/internal/validator"
	"net/http"
)

// Home displays the home page
func (app *application) Home(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "home", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// Register displays the registration page
func (app *application) Register(w http.ResponseWriter, r *http.Request) {

	userTypes, err := app.DB.GetAllUserTypes()
	if err != nil {
		app.errorLog.Println(err)
	}
	data := make(map[string]interface{})
	data["UserTypes"] = userTypes

	if err := app.renderTemplate(w, r, "register", &templateData{
		Data:      data,
		Validator: validator.New(),
	}); err != nil {
		app.errorLog.Println(err)
	}
}
