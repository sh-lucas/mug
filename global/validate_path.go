package global

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//go:embed .mugignore
var mugIgnore string

var ignorableGlobs []string

func init() {
	extra, err := os.ReadFile(".mugignore")
	if err == nil {
		mugIgnore += string(extra)
	}

	for _, toIgnore := range strings.Split(mugIgnore, "\n") {
		toIgnore = strings.TrimSpace(toIgnore)
		ignorableGlobs = append(ignorableGlobs, toIgnore)
	}
}

func validPath(path string) bool {
	path = filepath.Base(path)
	for _, glob := range ignorableGlobs {
		matched, err := filepath.Match(glob, path)
		if err != nil {
			log.Fatalf("invalid glob in mugignore")
		}
		if matched {
			Logf("Path %s not tracked because of glob %s\n", path, glob)
			return false
		}
	}
	return true
}
