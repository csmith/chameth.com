package achievements

import (
	"bytes"
	_ "embed"
	"html/template"
)

//go:embed *.gotpl
var templates string

var tmpl = template.Must(template.New("achievements.html.gotpl").Funcs(template.FuncMap{
	"formatDate": formatDate,
}).Parse(templates))

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func formatDate(t any) string {
	switch v := t.(type) {
	case interface{ Format(string) string }:
		return v.Format("2 Jan 2006")
	}
	return ""
}
