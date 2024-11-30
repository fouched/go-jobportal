package main

import (
	"database/sql"
	"errors"
	"github.com/fouched/go-jobportal/internal/models"
	"github.com/fouched/go-jobportal/internal/validator"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
)

// Home displays the home page
func (app *application) Home(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "home", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

// ShowRegister displays the registration page
func (app *application) ShowRegister(w http.ResponseWriter, r *http.Request) {

	userTypes, err := app.DB.GetAllUserTypes()
	if err != nil {
		app.errorLog.Println(err)
	}
	data := make(map[string]interface{})
	data["UserTypes"] = userTypes

	val := validator.New()
	// testing...
	//val.Check(false, "UserExists", "User already registered")
	//val.Check(false, "InvalidEmail", "Invalid email")

	if err := app.renderTemplate(w, r, "register", &templateData{
		Data:      data,
		Validator: val,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) RegisterNew(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")
	userTypeID, _ := strconv.Atoi(r.Form.Get("userTypeID"))

	//TODO form validation

	// check if email already exists
	_, err = app.DB.GetUserByEmail(email)
	ok := false
	switch {
	case errors.Is(err, sql.ErrNoRows):
		ok = true
	case err != nil:
		ok = false
	default:
		ok = false
	}

	if ok {
		userType := models.UserType{
			ID: userTypeID,
		}

		user := models.User{
			Email:    email,
			IsActive: true,
			UserType: &userType,
		}

		newHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			app.errorLog.Println(err)
			return
		}

		err = app.DB.AddUser(user, string(newHash))
		if err != nil {
			app.errorLog.Println(err)
			return
		}

		if err := app.renderTemplate(w, r, "dashboard", &templateData{
			Validator: validator.New(),
		}); err != nil {
			app.errorLog.Println(err)
		}
	} else {
		val := validator.New()
		val.Check(false, "UserExists", "User already registered")

		if err := app.renderTemplate(w, r, "register", &templateData{
			Validator: val,
		}); err != nil {
			app.errorLog.Println(err)
		}
	}
}
