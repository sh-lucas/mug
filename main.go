package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start listening for events.
	go watch(watcher)
	go waiter()

	// TODO: wait for signals
	<-make(chan struct{})
}

func watch(watcher *fsnotify.Watcher) {
	// Add current path.
	Add(watcher, ".")

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			info, err := os.Stat(event.Name)
			// errors if the file was deleted
			if err == nil && info.IsDir() {
				Add(watcher, event.Name)
			}

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
				Signals <- true
			}
		case err, ok := <-watcher.Errors:
			log.Println("error:", err)
			if !ok {
				return
			}
		}
	}
}

//go:embed mug.ignore
var mugIgnore string

func Add(watcher *fsnotify.Watcher, path string) {
	for _, ignore := range strings.Split(mugIgnore, "\n") {
		ignore = strings.TrimSpace(ignore)
		if path == ignore /* || strings.Contains(path, ignore) */ {
			fmt.Println("Ignoring path:", path, "due to ignore rule:", ignore)
			return // ignore this path
		}
	}

	if err := watcher.Add(path); err != nil {
		log.Println("Failed to add path:", path, err)
	} else {
		log.Println("tracking path", path)
	}
	files, err := os.ReadDir(path)
	if err != nil {
		log.Println("Failed to read directory:", path, err)
		return
	}
	for _, file := range files {
		fullPath := path + string(os.PathSeparator) + file.Name()
		if file.IsDir() {
			Add(watcher, fullPath)
		}
	}
}

var Signals = make(chan bool, 5)
var debouceTime = 200 * time.Millisecond

// waiter implements a simple debounce logic
// to avoid multiple rebuilds
// this means that if you spam ctrl + s, it will only
// rebuild after 200ms of the last signal
func waiter() {
	for {
		<-Signals // waits for the first signal
		timer := time.NewTimer(debouceTime)

	debounceLoop:
		for {
			select {
			case <-Signals:
				// if a new signal comes before the timer, resets
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(debouceTime)
			case <-timer.C:
				// if not, rebuilds already
				rebuild()
				break debounceLoop
			}
		}
	}
}

var runningAppProcess *os.Process

func rebuild() {
	// kills previous process
	if runningAppProcess != nil {
		_ = runningAppProcess.Kill()
	}

	log.Println("Rebuilding application...")
	runCmd := exec.Command("go", "run", ".")
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	if err := runCmd.Start(); err != nil {
		log.Printf("❌ Failed to start new process: %v", err)
		return
	}

	runningAppProcess = runCmd.Process
}
