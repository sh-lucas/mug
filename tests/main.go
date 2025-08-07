package main

import (
	cup "github.com/sh-lucas/mug/tests/mug_generated"
	cup_router "github.com/sh-lucas/mug/tests/mug_generated/router"
)

func main() {
	// example on how to use the cup package
	// this routes on port 8080 =)
	if cup.TEST != "TestValue1" {
		panic("TEST env is not set to 'TestValue1'")
	}

	cup_router.Route("8080")
}
