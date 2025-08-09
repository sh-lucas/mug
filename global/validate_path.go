package global

import (
	_ "embed"
	"log"
	"path/filepath"
	"strings"
)

//go:embed .mugignore
var mugIgnore string

// checks if the path is in mugignore and validates it
func ValidatePath(path string) bool {
	for _, toIgnore := range strings.Split(mugIgnore, "\n") {
		toIgnore = strings.TrimSpace(toIgnore)
		ignore, err := filepath.Match(toIgnore, path)
		if err != nil {
			log.Fatalf("Invalid Glob in mugignore")
		}
		if ignore {
			return false
		}
	}
	return true
}
