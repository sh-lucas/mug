package global

import (
	"log"
)

func Logf(pattern string, params ...any) {
	if pattern[len(pattern)-1] != '\n' {
		pattern += "/n"
	}
	if Config.Debug {
		log.Printf(pattern, params...)
	}
}
