package helpers

import (
	"fmt"
	"log"

	"github.com/sh-lucas/mug/internal/config"
)

func Logf(pattern string, params ...any) {
	if pattern[len(pattern)-1] != '\n' {
		pattern = fmt.Sprintf("%s\n", pattern)
	}
	if config.Global.Debug {
		log.Printf(pattern, params...)
	}
}
