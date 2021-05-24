package esbuild

import (
	"github.com/evanw/esbuild/pkg/api"
)

type Build struct {
	options api.BuildOptions
}

// WithWatch configures the Watch option on esbuild
func (b *Build) WithWatch(f func(result api.BuildResult)) *Build {
	b.options.Watch = &api.WatchMode{
		OnRebuild: f,
	}

	return b
}

// Run makes esbuild run the build process.
func (b *Build) Run() api.BuildResult {
	result := api.Build(b.options)

	return result
}

// NewBuild creates a new build struct.
func NewBuild(entryPoint string) Build {
	b := Build{
		options: api.BuildOptions{
			EntryPoints      : []string{entryPoint},
			Bundle           : true,
			MinifyWhitespace : false,
			MinifyIdentifiers: false,
			MinifySyntax     : false,
			Outfile          : "index.js",
			Engines          : []api.Engine{{Name: api.EngineChrome, Version: "91"}},
			Write            : false,
		},
	}

	return b
}
