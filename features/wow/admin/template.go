package admin

import (
	_ "embed"
	"net/http"

	adminTemplates "chameth.com/chameth.com/features/admin/templates"
)

//go:embed list-characters.html.gotpl
var listCharactersGotpl string

var listCharactersTemplate = adminTemplates.ParsePage(listCharactersGotpl)

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
