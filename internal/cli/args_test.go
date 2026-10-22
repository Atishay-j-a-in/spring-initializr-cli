package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		var out bytes.Buffer
		opts, err := ParseArgs([]string{"--url", "http://localhost:8080/", "--output", "out", "--verbose"}, &out, "test")
		if err != nil {
			t.Fatal(err)
		}
		if opts.BaseURL != "http://localhost:8080" || !opts.Verbose || opts.OutputDir == "" {
			t.Fatalf("unexpected options: %#v", opts)
		}
	})
	t.Run("unknown", func(t *testing.T) {
		var out bytes.Buffer
		_, err := ParseArgs([]string{"--wat"}, &out, "test")
		if err == nil || !strings.Contains(out.String(), "Usage:") {
			t.Fatalf("expected usage and error: %v, %q", err, out.String())
		}
	})
	t.Run("missing value", func(t *testing.T) {
		var out bytes.Buffer
		_, err := ParseArgs([]string{"--url"}, &out, "test")
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("help", func(t *testing.T) {
		var out bytes.Buffer
		opts, err := ParseArgs([]string{"--help"}, &out, "test")
		if err != nil || !opts.ShowHelp || !strings.Contains(out.String(), "Usage:") {
			t.Fatalf("unexpected: %#v %v %q", opts, err, out.String())
		}
	})
	t.Run("version", func(t *testing.T) {
		var out bytes.Buffer
		opts, err := ParseArgs([]string{"--version"}, &out, "test")
		if err != nil || !opts.ShowVersion || out.String() != "spring-create test\n" {
			t.Fatalf("unexpected: %#v %v %q", opts, err, out.String())
		}
	})
}
