package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
)

const DefaultURL = "https://start.spring.io"

type Options struct {
	BaseURL     string
	OutputDir   string
	NoColor     bool
	Verbose     bool
	ShowHelp    bool
	ShowVersion bool
}

func ParseArgs(args []string, out io.Writer, version string) (Options, error) {
	var opts Options
	fs := flag.NewFlagSet("spring-create", flag.ContinueOnError)
	var parseOutput bytes.Buffer
	fs.SetOutput(&parseOutput)
	fs.StringVar(&opts.BaseURL, "url", DefaultURL, "Initializr base URL")
	fs.StringVar(&opts.OutputDir, "output", "", "output parent directory")
	fs.BoolVar(&opts.NoColor, "no-color", false, "disable color output")
	fs.BoolVar(&opts.Verbose, "verbose", false, "print diagnostic context to stderr")
	fs.BoolVar(&opts.ShowHelp, "help", false, "print usage")
	fs.BoolVar(&opts.ShowHelp, "h", false, "print usage")
	fs.BoolVar(&opts.ShowVersion, "version", false, "print version")
	fs.Usage = func() { printUsage(out, fs) }

	if err := fs.Parse(args); err != nil {
		printUsage(out, fs)
		return Options{}, errors.New(strings.TrimSpace(parseOutput.String()))
	}
	if fs.NArg() != 0 {
		printUsage(out, fs)
		return Options{}, fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	if opts.ShowHelp {
		printUsage(out, fs)
		return opts, nil
	}
	if opts.ShowVersion {
		fmt.Fprintf(out, "spring-create %s\n", version)
		return opts, nil
	}

	u, err := url.Parse(opts.BaseURL)
	if err != nil || !u.IsAbs() || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return Options{}, fmt.Errorf("--url must be an absolute http or https URL")
	}
	opts.BaseURL = strings.TrimRight(u.String(), "/")
	if opts.OutputDir != "" {
		abs, err := filepath.Abs(opts.OutputDir)
		if err != nil {
			return Options{}, fmt.Errorf("invalid --output path: %w", err)
		}
		opts.OutputDir = abs
	}
	return opts, nil
}

func printUsage(out io.Writer, fs *flag.FlagSet) {
	fmt.Fprintln(out, "Usage: spring-create [options]")
	fmt.Fprintln(out, "\nInteractively create a Spring Boot project with Spring Initializr.")
	fmt.Fprintln(out, "\nOptions:")
	fs.SetOutput(out)
	fs.PrintDefaults()
}
