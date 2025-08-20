package watcher

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sh-lucas/mug/helpers"
)

type Task func() *exec.Cmd

var stopSig = make(chan os.Signal, 1)

var Signals = make(chan bool, 1)

var debouceTime = 350 * time.Millisecond

var running *exec.Cmd

func Start(task Task) {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	signal.Notify(stopSig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	// Start listening for events.
	go watch(watcher)
	go waiter(task)

	// wait for SIGINT/SIGTERM signals
	<-stopSig
	if running != nil && running.Process != nil {
		Kill()
	}
	err = syscall.Kill(0, syscall.SIGKILL)
	if err != nil {
		os.Exit(0)
	} else {
		log.Fatal(err)
	}
}

// looks for modifications in the current directory
func watch(watcher *fsnotify.Watcher) {
	// Add current path.
	Add(watcher, ".", 0)

	// rebuilds for the first time
	Signals <- true

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok { // watcher closed
				return
			}

			info, err := os.Stat(event.Name)
			if err == nil && info.IsDir() {
				if helpers.ValidPath(filepath.Base(info.Name())) {
					Add(watcher, event.Name, 1)
				}
			}

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
				select {
				case Signals <- true:
				default: // skips
				}
			}
		case err, ok := <-watcher.Errors:
			if err != nil {
				log.Println("Error in reloader:", err)
			}
			if !ok {
				return
			}
		}
	}
}

// Adds the current path to the watcher and
// recursively adds all subdirectories
func Add(watcher *fsnotify.Watcher, path string, depth int) {
	helpers.Walk(path, func(filepath string) {
		if err := watcher.Add(filepath); err != nil {
			log.Fatalf("Could not watch:\n %s\n err: \n%v", filepath, err)
		}
	})
}

// waiter implements a simple debounce logic
// to avoid multiple rebuilds
// this means that if you spam ctrl + s, it will only
// rebuild after 200ms of the last signal
func waiter(task Task) {
	for {
		<-Signals // waits for signals
		timer := time.NewTimer(debouceTime)

	debounceLoop:
		for {
			select {
			case <-Signals:
				// if a new signal comes before the timer, reset
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(debouceTime)
			case <-timer.C:
				Kill()
				fmt.Println(helpers.Blue + "> Rebuilding application" + helpers.Reset)
				running = task()
				clearChan(Signals)
				break debounceLoop
			}
		}
	}
}

// gracefully stop the running process
// it's patience only lasts for 3 seconds
func Kill() {
	if running != nil && running.Process != nil {
		err := syscall.Kill(-running.Process.Pid, syscall.SIGTERM)
		if err != nil {
			log.Println("Failed to kill process:", err)
		}

		done := make(chan error)
		go func() {
			done <- running.Wait()
		}()

		select {
		case <-time.After(3 * time.Second):
			log.Println("Process did not exit in time, killing it forcefully")
			err = syscall.Kill(-running.Process.Pid, syscall.SIGKILL)
			if err != nil {
				log.Fatalln("Failed to kill process forcefully:", err)
			}
		case <-done:
			// process exited gracefully
			// error ignored because it might
			// have exited after kill.
		}
	}
}

func clearChan[T any](c chan T) {
	for len(c) > 0 {
		<-c
	}
}
