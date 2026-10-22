package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"example.com/spring-create/internal/cli"
	"example.com/spring-create/internal/initializr"
	"example.com/spring-create/internal/project"
)

const (
	ExitSuccess  = 0
	ExitInternal = 1
	ExitInput    = 2
	ExitNetwork  = 3
	ExitFile     = 4
)

type CodedError struct {
	Code int
	Err  error
}

func (e *CodedError) Error() string { return e.Err.Error() }
func (e *CodedError) Unwrap() error { return e.Err }

func NewError(code int, format string, args ...any) error {
	return &CodedError{Code: code, Err: fmt.Errorf(format, args...)}
}

func ExitCode(err error) int {
	if err == nil || errors.Is(err, cli.ErrCanceled) || errors.Is(err, context.Canceled) {
		return ExitSuccess
	}
	var coded *CodedError
	if errors.As(err, &coded) {
		return coded.Code
	}
	return ExitInternal
}

func Run(ctx context.Context, args []string, in io.Reader, out, errOut io.Writer, version string) error {
	opts, err := cli.ParseArgs(args, out, version)
	if err != nil {
		return NewError(ExitInput, "%w", err)
	}
	if opts.ShowHelp || opts.ShowVersion {
		return nil
	}
	if !cli.IsInteractive(in) {
		return NewError(ExitInput, "interactive input is required; run spring-create from a terminal")
	}
	theme := cli.NewTheme(out, opts.NoColor)
	if opts.Verbose {
		fmt.Fprintf(errOut, "Fetching metadata from %s\n", opts.BaseURL)
	}
	client, err := initializr.NewClient(opts.BaseURL, "spring-create/"+version, initializr.MetadataTimeout)
	if err != nil {
		return NewError(ExitInput, "%w", err)
	}
	metadata, err := client.FetchMetadata(ctx)
	if err != nil {
		return NewError(ExitNetwork, "%v. Check your internet connection or use --url for a compatible Initializr service.", err)
	}
	config, err := cli.RunWizard(in, out, metadata, opts.OutputDir, theme)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, "\n"+theme.Accent("[1/3]")+" Downloading generated project...")
	zipPath, err := project.DownloadZIP(ctx, &http.Client{Timeout: project.GenerationTimeout}, client.GenerationURL(initializr.GenerationRequest{Type: config.Type, Language: config.Language, BootVersion: config.BootVersion, BaseDir: config.ArtifactID, GroupID: config.GroupID, ArtifactID: config.ArtifactID, Name: config.Name, Description: config.Description, PackageName: config.PackageName, Packaging: config.Packaging, JavaVersion: config.JavaVersion, Dependencies: config.Dependencies}), "spring-create/"+version)
	if err != nil {
		return NewError(ExitNetwork, "%w", err)
	}
	defer osRemove(zipPath)
	fmt.Fprintln(out, theme.Accent("[2/3]")+" Validating and safely extracting archive...")
	destination, err := project.ExtractZIP(zipPath, config.OutputParent, config.ArtifactID)
	if err != nil {
		return NewError(ExitFile, "%w", err)
	}
	abs, _ := filepath.Abs(destination)
	fmt.Fprintln(out, theme.Success("[3/3] Complete"))
	fmt.Fprintln(out, theme.Muted("----------------------------------------------------------------"))
	fmt.Fprintf(out, "%s\n  %s\n", theme.Success("Project created at:"), abs)
	if config.Type == "maven-project" {
		fmt.Fprintf(out, "Next: cd %q && .\\mvnw.cmd spring-boot:run\n", abs)
	} else {
		fmt.Fprintf(out, "Next: cd %q && .\\gradlew.bat bootRun\n", abs)
	}
	fmt.Fprintln(out, "A compatible JDK is required to build or run the generated project.")
	return nil
}

func osRemove(path string) { _ = os.Remove(path) }
