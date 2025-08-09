package global

import (
	"flag"
	"log"
)

var dbgInfo = flag.Bool("dbg", false, "Enables debug info")

func Logf(pattern string, params ...any) {
	if pattern[len(pattern)-1] != '\n' {
		pattern += "/n"
	}
	if *dbgInfo {
		log.Printf(pattern, params...)
	}
}
