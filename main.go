package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
)

var inputDir string

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <input_dir> <mount_point>", os.Args[0])
	}
	inputDir = os.Args[1]
	inputDir, err := filepath.Abs(inputDir)
	if err != nil {
		log.Fatalf("Error converting to absolute path: %v", err)
		return
	}
	mountPoint := os.Args[2]

	fs := NewPeopleFS()

	// Wait group to wait for goroutines to complete
	var wg sync.WaitGroup

	// Start the indexing goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := fs.Reindex(inputDir); err != nil {
			log.Fatalf("Reindex error: %v", err)
		}
	}()

	// Start the file watching goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		WatchInput(inputDir, fs)
	}()

	// Start the file system mount goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := MountFS(mountPoint, fs); err != nil {
			log.Fatalf("Mount error: %v", err)
		}
	}()

	// Handle system signals (SIGINT, SIGTERM)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// Gracefully handle termination and unmount FS
	go func() {
		<-sig
		log.Println("Received termination signal, unmounting filesystem...")
		// Unmount the filesystem here if needed
		if err := UnmountFS(mountPoint); err != nil {
			log.Printf("Failed to unmount filesystem: %v", err)
		} else {
			log.Println("Filesystem unmounted successfully.")
		}
		os.Exit(0)
	}()

	// Wait for all goroutines to complete
	wg.Wait()
}
