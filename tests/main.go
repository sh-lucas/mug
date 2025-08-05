package main

import (
	"log"
	"time"
)

func main() {
	log.Println("Test application running!")
	for {
		// Simulate some work
		log.Println("Working...")
		time.Sleep(2 * time.Second)
	}
}
