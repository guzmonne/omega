package esbuild

import (
	"github.com/evanw/esbuild/pkg/api"
)

type Build struct {
	options api.BuildOptions
	files		map[string][]byte
}

// WithWatch configures the Watch option on esbuild
func (b *Build) WithWatch(f func(result api.BuildResult)) *Build {
	b.options.Watch = &api.WatchMode{
		OnRebuild: f,
	}

	return b
}

// WithEntrypoints configures the entry points of the build process.
func (b *Build) WithEntrypoints(entryPoints []string) *Build {
	b.options.EntryPoints = entryPoints

	return b
}

// Run makes esbuild run the build process.
func (b *Build) Run() api.BuildResult {
	result := api.Build(b.options)

	return result
}

// Set stores the content inside the files map, referenced by the name as key.
func (b *Build) Set(name string, content []byte) *Build {
	b.files[name] = content

	return b
}

// Get gets the contents of a build file.
func (b *Build) Get(name string) []byte {
	content, ok := b.files[name]
	if !ok {
		return []byte("")
	}

	return content
}

// NewBuild creates a new build struct.
func NewBuild() *Build {
	b := &Build{
		options: api.BuildOptions{
			Bundle           : true,
			MinifyWhitespace : false,
			MinifyIdentifiers: false,
			MinifySyntax     : false,
			Outfile          : "index.js",
			Engines          : []api.Engine{{Name: api.EngineChrome, Version: "91"}},
			Incremental      : true,
			Write            : false,
		},
		files  : map[string][]byte{},
	}

	return b
}
