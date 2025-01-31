// Package watcher provides a simple mechanism to watch a file for changes.
package watcher

import (
	"fmt"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// Watcher is a structure to watch a file for changes and notify a channel.
type Watcher struct {
	filename    string
	fileChanges chan bool
}

// NewWatcher instantiates a new watcher with the provided filename and changes
// channel.
func NewWatcher(filename string, fileChanges chan bool) *Watcher {
	return &Watcher{
		filename:    filename,
		fileChanges: fileChanges,
	}
}

// Run starts the file watcher in a blocking fashion. This watches an entire
// directory and only notifies the channel when the specified file is changed.
// If there is an error, it's returned. It's up to the caller to respawn the
// watcher if it's desireable to keep watching.
//
// The reason it watches a directory is becausde some editers like VIM write
// to a swap file and recreate the original file. So we can't simply watch the
// original file, we have to watch the directory. This is also why we check both
// the WRITE and CREATE events since VIM will write to a swap and then create
// the file on save. VSCode does a WRITE and then a CHMOD, so tracking WRITE
// catches the changes for VSCode exactly once.
func (w *Watcher) Run() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error watching for changes: %w", err)
	}
	defer watcher.Close()

	watcher.Add(filepath.Dir(w.filename))

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("something is weird with the file watcher")
			}
			if event.Name == w.filename && (event.Op&(fsnotify.Write|fsnotify.Create)) != 0 {
				w.fileChanges <- true
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("something is weird with the file watcher around error handling")
			}
			return fmt.Errorf("error in file watcher: %w", err)
		}
	}
}
