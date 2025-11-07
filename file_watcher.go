package main

import (
	"context"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func isGoFile(filename string) bool {
	return filepath.Ext(filename) == ".go"
}

func addWatchRecursive(watcher *fsnotify.Watcher, rootpath string) error {
	return filepath.WalkDir(rootpath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}
			err = watcher.Add(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func watchFiles(
	ctx context.Context,
	dir string,
	fileChangeChan chan FileChangeMessage,
	startWatchingChan chan struct{},
) {
	select {
	case <-startWatchingChan:
	// Proceed
	case <-ctx.Done():
		return
	}
	watcher, err := fsnotify.NewWatcher()
	defer func() {
		err := watcher.Close()
		if err != nil {
			log.Print(err)
		}
	}()

	if err != nil {
		log.Print(err)
	}
	err = addWatchRecursive(watcher, dir)
	if err != nil {
		log.Print(err)
	}

	debounceChan := make(chan fsnotify.Event, 10)
	go debounceLoop(200*time.Millisecond, debounceChan, func(_ fsnotify.Event) {
		fileChangeChan <- FileChangeMessage{}
	})

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if isTrackedChangeEvent(event) && filepath.Ext(event.Name) == ".go" {
				// fmt.Println(event.String())
				debounceChan <- event
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println(err)
		}
	}
}

func debounceLoop(interval time.Duration, input chan fsnotify.Event, callback func(event fsnotify.Event)) {
	var event fsnotify.Event
	timer := time.NewTimer(interval)
	<-timer.C

	for {
		select {
		case event = <-input:
			// fmt.Println("======= resetting debounce timer")
			timer.Reset(interval)
		case <-timer.C:
			// fmt.Println("===== timeout reached:")
			// fmt.Println("    ", event.String())
			callback(event)
		}
	}
}

func isTrackedChangeEvent(event fsnotify.Event) bool {
	return event.Has(fsnotify.Create) ||
		event.Has(fsnotify.Remove) ||
		event.Has(fsnotify.Write) ||
		event.Has(fsnotify.Rename)
}
