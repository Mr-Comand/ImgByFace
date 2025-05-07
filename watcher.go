package main

import (
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
)

func WatchInput(inputDir string, pfs *PeopleFS) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(inputDir)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create ||
				event.Op&fsnotify.Remove == fsnotify.Remove {
				log.Println("Detected change:", event)
				time.Sleep(500 * time.Millisecond) // debounce
				pfs.Reindex(inputDir)
			}
		case err := <-watcher.Errors:
			log.Println("Watcher error:", err)
		}
	}
}
