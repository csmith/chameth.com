package assets

import (
	"bytes"
	"flag"
	"fmt"
	"hash/crc32"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

var (
	styleDate = flag.String("style-datetime", "", "Date/time to fake for stylesheet generation purposes")
)

type stylesheet struct {
	path    string
	content []byte
}

var stylesheets []stylesheet
var compiledSheet = &stylesheet{}

func RegisterStylesheet(path string, content []byte) {
	stylesheets = append(stylesheets, stylesheet{
		path:    path,
		content: content,
	})
}

func init() {
	entries, err := fs.ReadDir(Stylesheets, "stylesheet")
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".css") {
			b, err := fs.ReadFile(Stylesheets, filepath.Join("stylesheet", entry.Name()))
			if err != nil {
				panic(err)
			}

			RegisterStylesheet(entry.Name(), b)
		}
	}
}

func UpdateStylesheet() error {
	activeMoods, err := activeMoods()
	if err != nil {
		return err
	}

	includes := append(stylesheets, activeMoods...)

	builder := new(bytes.Buffer)
	for i := range includes {
		fmt.Fprintf(builder, "\n\n/* =========================== %s ========================== */\n\n", includes[i].path)
		builder.Write(includes[i].content)
	}

	hasher := crc32.NewIEEE()
	if _, err := hasher.Write(builder.Bytes()); err != nil {
		return err
	}

	compiledSheet = &stylesheet{
		content: builder.Bytes(),
		path:    fmt.Sprintf("global-%x.css", hasher.Sum(nil)),
	}

	return nil
}

func Stylesheet() []byte {
	return compiledSheet.content
}

func StylesheetPath() string {
	return compiledSheet.path
}

type mood struct {
	include string
	test    func(time.Time) bool
}

var moods = []mood{
	{
		include: "moods/birthday.css",
		test: func(t time.Time) bool {
			return t.Month() == 10 && t.Day() == 22
		},
	},
	{
		include: "moods/christmas.css",
		test: func(t time.Time) bool {
			return t.Month() == 12
		},
	},
	{
		include: "moods/halloween.css",
		test: func(t time.Time) bool {
			return t.Month() == 10 && t.Day() >= 24
		},
	},
}

func activeMoods() ([]stylesheet, error) {
	date := time.Now()
	if *styleDate != "" {
		date, _ = time.ParseInLocation("2006-01-02T15:04:05", *styleDate, date.Location())
	}

	var active []stylesheet
	for _, mood := range moods {
		if mood.test(date) {
			b, err := fs.ReadFile(Moods, mood.include)
			if err != nil {
				return nil, err
			}

			active = append(active, stylesheet{
				path:    mood.include,
				content: b,
			})
		}
	}

	return active, nil
}
