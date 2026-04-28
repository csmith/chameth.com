package admin

import (
	"embed"
	"html/template"
	"net/http"

	adminTemplates "chameth.com/chameth.com/admin/templates"
)

//go:embed *.gotpl
var templates embed.FS

var listCharactersTemplate = template.Must(
	template.Must(
		template.New("page.html.gotpl").ParseFS(
			adminTemplates.Templates,
			"page.html.gotpl",
		),
	).ParseFS(
		templates,
		"list-characters.html.gotpl",
	),
)

type ListCharactersData struct {
	adminTemplates.PageData
	Characters []CharacterSummary
}

type CharacterSummary struct {
	ID            int
	CharacterName string
	RealmName     string
	Race          string
	Class         string
	Spec          string
	Gender        string
	Faction       string
	UpdatedAt     string
}

func renderListCharacters(w http.ResponseWriter, data ListCharactersData) error {
	return listCharactersTemplate.Execute(w, data)
}
