package main

import (
	"embed"
	"fmt"
	"github.com/fouched/go-jobportal/internal/validator"
	"html/template"
	"net/http"
	"strings"
)

// TemplateData holds data sent from handlers to templates
type templateData struct {
	StringMap  map[string]string
	IntMap     map[string]int
	FloatMap   map[string]float32
	BoolMap    map[string]bool
	Data       map[string]interface{}
	CSRFToken  string
	Success    string
	Warning    string
	Error      string
	AuthLevel  int
	UserID     int
	UserName   string
	CSSVersion string
	Validator  *validator.Validator
}

// with the go embed directive below we can compile
// the templates with the application in a single binary
//
//go:embed templates
var templateFS embed.FS

var functions = template.FuncMap{
	"unescapeHTML": unescapeHTML,
}

func unescapeHTML(s string) template.HTML {
	return template.HTML(s)
}

func (app *application) renderTemplate(w http.ResponseWriter, r *http.Request, page string, td *templateData, partials ...string) error {
	var t *template.Template
	var err error
	templateToRender := fmt.Sprintf("templates/%s.page.gohtml", page)

	_, templateInMap := app.templateCache[templateToRender]

	if templateInMap {
		t = app.templateCache[templateToRender]
	} else {
		t, err = app.parseTemplate(partials, page, templateToRender)
		if err != nil {
			app.errorLog.Println(err)
			return err
		}
	}

	if td == nil {
		td = &templateData{}
	}

	td = app.addDefaultData(td, r)

	err = t.Execute(w, td)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	return nil
}

func (app *application) parseTemplate(partials []string, page, templateToRender string) (*template.Template, error) {
	var t *template.Template
	var err error

	// build partials
	if len(partials) > 0 {
		for i, x := range partials {
			partials[i] = fmt.Sprintf("templates/%s.partial.gohtml", x)
		}
	}

	if len(partials) > 0 {
		t, err = template.New(fmt.Sprintf("%s.page.gohtml", page)).Funcs(functions).
			ParseFS(templateFS, "templates/base.layout.gohtml", strings.Join(partials, ","), templateToRender)
	} else {
		t, err = template.New(fmt.Sprintf("%s.page.gohtml", page)).Funcs(functions).
			ParseFS(templateFS, "templates/base.layout.gohtml", templateToRender)
	}

	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}

	app.templateCache[templateToRender] = t

	return t, nil
}

func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	if app.Session.Exists(r.Context(), "userID") {
		userId := app.Session.GetInt(r.Context(), "userID")
		authLevel := app.Session.GetInt(r.Context(), "userTypeID")

		td.AuthLevel = authLevel
		td.UserID = userId
		td.UserName = app.Session.GetString(r.Context(), "userName")

		if authLevel == 1 { // recruiter
			p, err := app.DB.GetRecruiterProfile(userId)
			if err != nil {
				app.errorLog.Println(err)
			}
			if p.FirstName != "" && p.LastName != "" {
				td.Data["FullName"] = p.FirstName + " " + p.LastName
			}
			if p.ProfilePhoto != "" {
				td.Data["ProfilePhoto"] = p.ProfilePhoto
			}
		} else { // jobseeker
			p, err := app.DB.GetJobSeekerProfile(userId)
			if err != nil {
				app.errorLog.Println(err)
			}
			if p.FirstName != "" && p.LastName != "" {
				td.Data["FullName"] = p.FirstName + " " + p.LastName
			}
			if p.ProfilePhoto != "" {
				td.Data["ProfilePhoto"] = p.ProfilePhoto
			}
		}

	} else {
		td.AuthLevel = 0
	}
	return td
}
