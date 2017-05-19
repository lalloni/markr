package options

import (
	"context"
	"flag"
)

type Options struct {
	InputFile   string
	OutputFile  string
	Cache       bool
	Verbose     bool
	Diagrams    string
	DiagramsDPI int
	Usage       bool
}

var options Options

const optionsKey = "options"

func ConfigureFlags() {
	flag.StringVar(&options.InputFile, "in", "", "Markdown input `file`")
	flag.StringVar(&options.OutputFile, "out", "", "PDF output `file`")
	flag.BoolVar(&options.Cache, "cache", false, "Cache generated intermediate results")
	flag.BoolVar(&options.Verbose, "verbose", false, "Be verbose")
	flag.StringVar(&options.Diagrams, "diagrams", "pdf", "Diagrams `format`: \"eps\" or \"pdf\"")
	flag.IntVar(&options.DiagramsDPI, "resolution", 300, "Diagrams `dpi` resolution")
	flag.BoolVar(&options.Usage, "help", false, "Show this help")
}

func WithOptions(ctx context.Context) context.Context {
	return context.WithValue(ctx, optionsKey, &options)
}

func Get(ctx context.Context) *Options {
	return ctx.Value(optionsKey).(*Options)
}
