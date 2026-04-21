package wowchar

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/features/wow"
)

//go:embed *.gotpl
var templates string

var tmpl = template.Must(template.New("wowchar.html.gotpl").Parse(templates))

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("wowchar requires 2 arguments (realm character)")
	}

	c, err := wow.GetCharacter(ctx, args[0], args[1])
	if err != nil {
		return "", fmt.Errorf("failed to get character: %w", err)
	}

	itemLevel := ""
	if c.EquippedItemLevel != nil {
		itemLevel = fmt.Sprintf("%d", *c.EquippedItemLevel)
	}

	return renderTemplate(Data{
		Name:              c.CharacterName,
		Realm:             c.RealmName,
		Level:             c.Level,
		Spec:              c.Spec,
		Class:             c.Class,
		Race:              c.Race,
		Gender:            c.Gender,
		EquippedItemLevel: itemLevel,
	})
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
