package cli

import (
	"io"
	"os"
)

const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiDim    = "\x1b[2m"
	ansiGreen  = "\x1b[92m"
	ansiYellow = "\x1b[93m"
	ansiRed    = "\x1b[91m"
)

// Theme adds optional ANSI styling while preserving readable plain-text output.
type Theme struct{ Enabled bool }

func NewTheme(out io.Writer, noColor bool) Theme {
	if noColor || os.Getenv("NO_COLOR") != "" {
		return Theme{}
	}
	file, ok := out.(terminalFile)
	if !ok {
		return Theme{}
	}
	info, err := file.Stat()
	return Theme{Enabled: err == nil && info.Mode()&os.ModeCharDevice != 0}
}

func (t Theme) Title(value string) string   { return t.style(ansiBold+ansiGreen, value) }
func (t Theme) Accent(value string) string  { return t.style(ansiGreen, value) }
func (t Theme) Muted(value string) string   { return t.style(ansiDim, value) }
func (t Theme) Warning(value string) string { return t.style(ansiYellow, value) }
func (t Theme) Error(value string) string   { return t.style(ansiRed, value) }
func (t Theme) Success(value string) string { return t.style(ansiBold+ansiGreen, value) }

func (t Theme) style(code, value string) string {
	if !t.Enabled {
		return value
	}
	return code + value + ansiReset
}

type terminalFile interface{ Stat() (os.FileInfo, error) }
