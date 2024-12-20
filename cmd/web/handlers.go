package main

import (
	"database/sql"
	"errors"
	"github.com/fouched/go-jobportal/internal/models"
	"github.com/fouched/go-jobportal/internal/validator"
	"github.com/fouched/toolkit/v2"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
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

		if err := app.renderTemplate(w, r, "login", &templateData{
			Validator: validator.New(),
		}); err != nil {
			app.errorLog.Println(err)
		}
	} else {
		val := validator.New()
		val.Check(false, "UserExists", "User already registered")

		userTypes, err := app.DB.GetAllUserTypes()
		if err != nil {
			app.errorLog.Println(err)
		}
		data := make(map[string]interface{})
		data["UserTypes"] = userTypes

		if err := app.renderTemplate(w, r, "register", &templateData{
			Validator: val,
			Data:      data,
		}); err != nil {
			app.errorLog.Println(err)
		}
	}
}

func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "login", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) LoginPost(w http.ResponseWriter, r *http.Request) {
	// good practice to renew token on login
	app.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	email := r.Form.Get("username")
	password := r.Form.Get("password")

	id, userTypeId, err := app.DB.Authenticate(email, password)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	app.Session.Put(r.Context(), "userID", id)
	app.Session.Put(r.Context(), "userTypeID", userTypeId)
	app.Session.Put(r.Context(), "userName", email)
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)

}

func (app *application) Logout(w http.ResponseWriter, r *http.Request) {
	app.Session.Destroy(r.Context())
	// good practice to renew token on login
	app.Session.RenewToken(r.Context())

	if err := app.renderTemplate(w, r, "home", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) Dashboard(w http.ResponseWriter, r *http.Request) {
	intMap := make(map[string]int)
	intMap["ShowNav"] = 1

	// we will only hit this route if authenticated, load the appropriate profile
	data := make(map[string]interface{})
	id := app.Session.GetInt(r.Context(), "userID")
	rp, err := app.DB.GetRecruiterProfile(id)
	if err != nil {
		app.errorLog.Println(err)
	}
	if rp.FirstName != "" && rp.LastName != "" {
		data["FullName"] = rp.FirstName + " " + rp.LastName
	}

	if rp.ProfilePhoto != "" {
		data["PhotosImagePath"] = "/uploads/" + rp.ProfilePhoto
	}

	jp, err := app.DB.GetRecruiterJobPosts(id)
	if err != nil {
		app.errorLog.Println(err)
	}
	data["JobPosts"] = jp

	if err := app.renderTemplate(w, r, "dashboard", &templateData{
		IntMap: intMap,
		Data:   data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) RecruiterProfile(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	rp, err := app.DB.GetRecruiterProfile(app.Session.GetInt(r.Context(), "userID"))
	if err != nil {
		app.errorLog.Println(err)
	}

	data["FirstName"] = rp.FirstName
	data["LastName"] = rp.LastName
	data["City"] = rp.City
	data["State"] = rp.State
	data["Country"] = rp.Country
	data["Company"] = rp.Company

	if err := app.renderTemplate(w, r, "recruiter-profile", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) RecruiterProfileUpdate(w http.ResponseWriter, r *http.Request) {

	t := toolkit.Tools{
		MaxFileSize:      5 * 1024 * 1024 * 1024,
		AllowedFileTypes: []string{"image/jpeg", "image/png", "image/gif"},
	}

	files, err := t.UploadFiles(r, "./uploads")
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// UploadFiles also parsed the form, so it is available to us
	rp := models.RecruiterProfile{
		UserAccountID: app.Session.GetInt(r.Context(), "userID"),
		FirstName:     r.Form.Get("firstName"),
		LastName:      r.Form.Get("lastName"),
		City:          r.Form.Get("city"),
		State:         r.Form.Get("state"),
		Country:       r.Form.Get("country"),
		Company:       r.Form.Get("company"),
	}

	err = app.DB.UpdateRecruiterProfile(rp)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// only update profile photo data if the user specified a file
	if len(files) > 0 {
		// delete previous profile photo
		rp, _ := app.DB.GetRecruiterProfile(app.Session.GetInt(r.Context(), "userID"))
		if rp.ProfilePhoto != "" {
			_ = os.Remove("./uploads/" + rp.ProfilePhoto)
		}

		rp.ProfilePhoto = files[0].NewFileName
		err = app.DB.UpdateRecruiterProfilePhoto(rp)
		if err != nil {
			app.errorLog.Println(err)
			return
		}
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func (app *application) JobPostAdd(w http.ResponseWriter, r *http.Request) {

	data := make(map[string]interface{})
	data["JobDetails"] = models.JobPost{ID: 0}

	if err := app.renderTemplate(w, r, "job-post", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) JobPostEdit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	jobID, _ := strconv.Atoi(id)

	jd, err := app.DB.GetJob(jobID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["JobDetails"] = jd

	if err := app.renderTemplate(w, r, "job-post", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) JobPostSave(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	id, _ := strconv.Atoi(r.Form.Get("jobPostId"))
	jp := models.JobPost{
		PostedById:  app.Session.GetInt(r.Context(), "userID"),
		ID:          id,
		Description: r.Form.Get("descriptionOfJob"),
		JobTitle:    r.Form.Get("jobTitle"),
		JobType:     r.Form.Get("jobType"),
		Salary:      r.Form.Get("salary"),
		Remote:      r.Form.Get("remote"),
		Location: models.JobLocation{
			City:    r.Form.Get("city"),
			State:   r.Form.Get("state"),
			Country: r.Form.Get("country"),
		},
		Company: models.JobCompany{
			Name: r.Form.Get("companyName"),
			Logo: "",
		},
	}

	err = app.DB.SaveJobPost(jp)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func (app *application) JobDetails(w http.ResponseWriter, r *http.Request) {
	intMap := make(map[string]int)
	intMap["ShowNav"] = 1

	id := chi.URLParam(r, "id")
	jobID, _ := strconv.Atoi(id)

	jd, err := app.DB.GetJob(jobID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["JobDetails"] = jd

	if err := app.renderTemplate(w, r, "job-details", &templateData{
		IntMap: intMap,
		Data:   data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) JobSeekerProfile(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "job-seeker-profile", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}
