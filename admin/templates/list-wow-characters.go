package templates

import (
	"html/template"
	"net/http"
)

var listWowCharactersTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"list-wow-characters.html.gotpl",
		),
)

type ListWowCharactersData struct {
	PageData
	Characters []WowCharacterSummary
}

type WowCharacterSummary struct {
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

func RenderListWowCharacters(w http.ResponseWriter, data ListWowCharactersData) error {
	return listWowCharactersTemplate.Execute(w, data)
}
