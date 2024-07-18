package handlers

import (
	"github.com/fouched/go-jobportal/internal/models"
	"github.com/fouched/go-jobportal/internal/render"
	"net/http"
)

func (a *HandlerConfig) Home(w http.ResponseWriter, r *http.Request) {

	templates := []string{"/pages/index.gohtml"}

	render.Templates(w, r, templates, true, &models.TemplateData{})
}
