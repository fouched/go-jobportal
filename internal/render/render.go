package render

import (
	"fmt"
	"github.com/fouched/go-jobportal/internal/config"
	"github.com/fouched/go-jobportal/internal/models"
	"html/template"
	"net/http"
)

var app *config.AppConfig
var pathToTemplates = "./templates"

// NewRenderer sets the config for the template package
func NewRenderer(a *config.AppConfig) {
	app = a
}

// Templates can render multiple templates. "Parent" templates should be defined first
func Templates(w http.ResponseWriter, r *http.Request, tmpl []string, addBaseTemplate bool, td *models.TemplateData) {

	for i, t := range tmpl {
		tmpl[i] = pathToTemplates + t
	}

	if addBaseTemplate {
		tmpl = append(tmpl, pathToTemplates+"/components/alert.gohtml", pathToTemplates+"/base.layout.gohtml")
	}

	td = AddDefaultData(td, r)
	parsedTemplate, _ := template.ParseFiles(tmpl...)
	err := parsedTemplate.Execute(w, td)
	if err != nil {
		fmt.Println("Error parsing template", err)
		return
	}

}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {

	//if helpers.IsAuthenticated(r) {
	//	td.IsAuthenticated = 1
	//}

	return td
}
