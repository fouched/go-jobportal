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
	"strings"
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

func (app *application) RegisterSave(w http.ResponseWriter, r *http.Request) {
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
	userId := app.Session.GetInt(r.Context(), "userID")
	userTypeID := app.Session.GetInt(r.Context(), "userTypeID")

	if userTypeID == 1 {
		p, err := app.DB.GetRecruiterProfile(userId)
		if err != nil {
			app.errorLog.Println(err)
		}
		if p.FirstName != "" && p.LastName != "" {
			data["FullName"] = p.FirstName + " " + p.LastName
		}
		if p.ProfilePhoto != "" {
			data["ProfilePhoto"] = p.ProfilePhoto
		}

		jp, err := app.DB.GetRecruiterJobPosts(userId)
		if err != nil {
			app.errorLog.Println(err)
		}
		data["JobPosts"] = jp
	} else {
		p, err := app.DB.GetJobSeekerProfile(userId)
		if err != nil {
			app.errorLog.Println(err)
		}
		if p.FirstName != "" && p.LastName != "" {
			data["FullName"] = p.FirstName + " " + p.LastName
		}
		if p.ProfilePhoto != "" {
			data["ProfilePhoto"] = p.ProfilePhoto
		}

		if r.Method == "POST" {
			sc, jp, err := app.DB.SearchJobPosts(r)
			if err != nil {
				app.errorLog.Println(err)
			}
			data["SearchCriteria"] = sc
			data["JobPosts"] = jp
		}
	}

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

func (app *application) RecruiterProfileSave(w http.ResponseWriter, r *http.Request) {

	userId := app.Session.GetInt(r.Context(), "userID")
	uploadDir := "./uploads/recruiter/" + strconv.Itoa(userId) + "/"

	t := toolkit.Tools{
		MaxFileSize:      5 * 1024 * 1024 * 1024,
		AllowedFileTypes: []string{"image/jpeg", "image/png", "image/gif", "image/bmp"},
	}

	// add current directory for the upload
	files, err := t.UploadFiles(r, uploadDir)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// UploadFiles also parsed the form, so it is available to us
	rp := models.RecruiterProfile{
		UserAccountID: userId,
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
		rp, _ := app.DB.GetRecruiterProfile(userId)
		if rp.ProfilePhoto != "" {
			_ = os.Remove("." + rp.ProfilePhoto)
		}

		rp.ProfilePhoto = strings.TrimPrefix(uploadDir, ".") + files[0].NewFileName
		err = app.DB.UpdateRecruiterProfilePhoto(rp)
		if err != nil {
			app.errorLog.Println(err)
			return
		}
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func (app *application) JobPost(w http.ResponseWriter, r *http.Request) {

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
	data := make(map[string]interface{})
	sp, err := app.DB.GetJobSeekerProfile(app.Session.GetInt(r.Context(), "userID"))
	if err != nil {
		app.errorLog.Println(err)
	}

	data["FirstName"] = sp.FirstName
	data["LastName"] = sp.LastName
	data["City"] = sp.City
	data["State"] = sp.State
	data["Country"] = sp.Country
	data["EmploymentType"] = sp.EmploymentType
	data["WorkAuthorization"] = sp.WorkAuthorization
	data["Skills"] = sp.Skills

	if err := app.renderTemplate(w, r, "job-seeker-profile", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) JobSeekerProfileSave(w http.ResponseWriter, r *http.Request) {

	userId := app.Session.GetInt(r.Context(), "userID")
	uploadDir := "./uploads/jobseeker/" + strconv.Itoa(userId) + "/"

	t := toolkit.Tools{
		MaxFileSize: 5 * 1024 * 1024 * 1024,
		AllowedFileTypes: []string{"image/jpeg", "image/png", "image/gif", "image/bmp",
			"application/pdf",
			"application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"application/vnd.ms-excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			"application/vnd.ms-powerpoint", "application/vnd.openxmlformats-officedocument.presentationml.presentation",
			"application/vnd.oasis.opendocument.text", "application/vnd.oasis.opendocument.spreadsheet", "application/vnd.oasis.opendocument.presentation"},
	}

	// add current directory for the upload
	files, err := t.UploadFiles(r, uploadDir)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// UploadFiles also parsed the form, so it is available to us
	// parse the skills
	var skills []models.Skill
	skillNames := r.Form["skillName"]
	skillExperienceLevel := r.Form["skillExperienceLevel"]
	skillYearsOfExperience := r.Form["skillYearsOfExperience"]
	for i, skillName := range skillNames {
		skill := models.Skill{
			UserAccountID:     userId,
			Name:              skillName,
			ExperienceLevel:   skillExperienceLevel[i],
			YearsOfExperience: skillYearsOfExperience[i],
		}
		skills = append(skills, skill)
	}

	sp := models.JobSeekerProfile{
		UserAccountID:     userId,
		FirstName:         r.Form.Get("firstName"),
		LastName:          r.Form.Get("lastName"),
		City:              r.Form.Get("city"),
		State:             r.Form.Get("state"),
		Country:           r.Form.Get("country"),
		WorkAuthorization: r.Form.Get("workAuthorization"),
		EmploymentType:    r.Form.Get("employmentType"),
		Skills:            skills,
	}

	err = app.DB.UpdateJobSeekerProfile(sp)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// only update profile photo data if the user specified a file
	if len(files) > 0 {
		// delete previous profile photo
		sp, _ := app.DB.GetJobSeekerProfile(userId)
		if sp.ProfilePhoto != "" {
			_ = os.Remove("." + sp.ProfilePhoto)
		}
		if sp.Resume != "" {
			_ = os.Remove("." + sp.Resume)
		}

		if files[0].Key == "profileImage" {
			sp.ProfilePhoto = strings.TrimPrefix(uploadDir, ".") + files[0].NewFileName
			sp.Resume = strings.TrimPrefix(uploadDir, ".") + files[1].NewFileName
		} else {
			sp.ProfilePhoto = strings.TrimPrefix(uploadDir, ".") + files[1].NewFileName
			sp.Resume = strings.TrimPrefix(uploadDir, ".") + files[0].NewFileName
		}

		err = app.DB.UpdateJobSeekerUploads(sp)
		if err != nil {
			app.errorLog.Println(err)
			return
		}
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}
