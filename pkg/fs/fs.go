package fs

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"time"
)

type FS struct {
	Root					string
	Paths 				map[string]time.Time
	Modified 			bool
	LastModified	time.Time
}

func (f *FS) Walk(fn func(path string, info os.FileInfo) error) error {
	return filepath.Walk(f.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fromNodeModules, err := regexp.MatchString(`node_modules`, path)
		if err != nil {
			return err
		}
		if fromNodeModules || info.IsDir() {
			return nil
		}
		return fn(path, info)
	})
}

// Refesh checks the fs from the root to see if updated where produced.
func (f *FS) Refresh() error {
	paths := map[string]time.Time{}

	if err := f.Walk(func(path string, info os.FileInfo) error {
		paths[path] = info.ModTime()
		return nil
	}); err != nil {
		return err
	}

	f.Modified 			= !reflect.DeepEqual(f.Paths, paths)
	f.LastModified 	= time.Now()
	f.Paths					= paths

	return nil
}

// newFS returns the initial state of the fs represented as a map of each file
// whose value represent the last modified date.
func NewFS(root string) (FS, error) {
	var paths map[string]time.Time

	fs := FS{
		Root					: root,
		Paths					: paths,
		Modified			: false,
		LastModified	: time.Now(),
	}

	if err := fs.Refresh(); err != nil {
		return fs, err
	}

	return fs, nil
}