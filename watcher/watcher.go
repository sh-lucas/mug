package watcher

import (
	_ "embed"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

var stopSig = make(chan os.Signal, 1)

func Start() {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	signal.Notify(stopSig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	// Start listening for events.
	go watch(watcher)
	go waiter()

	// wait for SIGINT/SIGTERM signals
	<-stopSig
	if running != nil && running.Process != nil {
		Kill()
	}
	err = syscall.Kill(0, syscall.SIGKILL)
	if err != nil {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

// looks for modifications in the current directory
func watch(watcher *fsnotify.Watcher) {
	// Add current path.
	Add(watcher, ".")

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
				Add(watcher, event.Name)
			}

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
				Signals <- true
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

//go:embed mug.ignore
var mugIgnore string

// Adds the current path to the watcher and
// recursively adds all subdirectories
func Add(watcher *fsnotify.Watcher, path string) {
	for _, ignore := range strings.Split(mugIgnore, "\n") {
		ignore = strings.TrimSpace(ignore)
		if path == ignore {
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
				log.Println("Failed to kill process forcefully:", err)
			}
		case <-done:
			// process exited gracefully
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

// a whole terminal for all the processes
var running *exec.Cmd

// kills the process and rebuilds the application
func rebuild() {
	// kills previous process
	Kill()

	log.Println("Rebuilding application...")
	cmd := exec.Command("go", "run", ".")
	// outputs
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// groups all processes and the current application
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		log.Printf("❌ Failed to start new process: %v", err)
		return
	}

	running = cmd
}
