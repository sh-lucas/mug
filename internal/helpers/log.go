package helpers

import (
	"fmt"
	"log"
)

func Logf(pattern string, params ...any) {
	if pattern[len(pattern)-1] != '\n' {
		pattern = fmt.Sprintf("%s\n", pattern)
	}
	if Config.Debug {
		log.Printf(pattern, params...)
	}
}
