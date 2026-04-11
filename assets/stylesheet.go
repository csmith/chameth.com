package assets

import (
	"flag"
	"fmt"
	"hash/crc32"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

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

var compiledSheet string
var compiledSheetPath string

var (
	styleDate = flag.String("style-datetime", "", "Date/time to fake for stylesheet generation purposes")
)

// UpdateStylesheet compiles all CSS files into a single stylesheet
// based on the current date or the date specified via -style-datetime flag.
func UpdateStylesheet() error {
	filesystem, err := fs.Sub(Stylesheets, filepath.Join("stylesheet"))
	if err != nil {
		return err
	}

	date := time.Now()
	if *styleDate != "" {
		date, _ = time.ParseInLocation("2006-01-02T15:04:05", *styleDate, date.Location())
	}

	entries, err := fs.ReadDir(filesystem, ".")
	if err != nil {
		return err
	}

	var includes []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".css") {
			includes = append(includes, entry.Name())
		}
	}
	for _, mood := range moods {
		if mood.test(date) {
			includes = append(includes, mood.include)
		}
	}

	builder := &strings.Builder{}
	for i := range includes {
		b, err := fs.ReadFile(filesystem, includes[i])
		if err != nil {
			return err
		}

		fmt.Fprintf(builder, "\n\n/* =========================== %s ========================== */\n\n", includes[i])
		builder.Write(b)
	}

	compiledSheet = builder.String()

	hasher := crc32.NewIEEE()
	if _, err := hasher.Write([]byte(compiledSheet)); err != nil {
		return err
	}
	compiledSheetPath = fmt.Sprintf("global-%x.css", hasher.Sum(nil))
	return nil
}

// GetStylesheet returns the compiled stylesheet content.
func GetStylesheet() string {
	return compiledSheet
}

// GetStylesheetPath returns the path of the compiled stylesheet.
func GetStylesheetPath() string {
	return compiledSheetPath
}
