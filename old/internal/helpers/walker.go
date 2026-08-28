package helpers

import (
	"log"
	"os"
	"path/filepath"
)

var maxDepth = 10

// recursively walks inside a folder recursively executing the forEach function
func Walk(path string, forEach func(entry string)) {
	walkRec(path, forEach, 0)
}

func walkRec(path string, forEach func(entry string), depth int) {
	if depth > maxDepth {
		log.Fatalf("Too deep folders; path: %s", path)
	}
	if !ValidPath(path) {
		Logf("Ignoring path %s", path)
		return
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		Logf("Invalid entry. Error: %v", err)
	}

	forEach(path)

	for _, entry := range entries {
		if entry.IsDir() {
			newPath := filepath.Join(path, entry.Name())
			walkRec(newPath, forEach, depth+1)
		}
	}
}
