package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var ErrCanceled = errors.New("operation canceled")

const rule = "----------------------------------------------------------------"

type Prompter struct {
	in    *bufio.Reader
	out   io.Writer
	theme Theme
}

func NewPrompter(in io.Reader, out io.Writer, theme Theme) *Prompter {
	return &Prompter{in: bufio.NewReader(in), out: out, theme: theme}
}

func PrintBanner(out io.Writer, theme Theme) {
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, theme.Accent("+--------------------------------------------------------------+"))
	fmt.Fprintln(out, theme.Title("|                        SPRING CREATE                         |"))
	fmt.Fprintln(out, theme.Muted("|             A guided Spring Boot project setup               |"))
	fmt.Fprintln(out, theme.Accent("+--------------------------------------------------------------+"))
	fmt.Fprintln(out, "")
}

func PrintSection(out io.Writer, theme Theme, title string) {
	fmt.Fprintf(out, "\n%s\n%s\n", theme.Title(title), theme.Muted(rule))
}

func (p *Prompter) Text(label, defaultValue string, required bool) (string, error) {
	for {
		if defaultValue == "" {
			fmt.Fprintf(p.out, "%s %s: ", p.theme.Accent("?"), label)
		} else {
			fmt.Fprintf(p.out, "%s %s %s: ", p.theme.Accent("?"), label, p.theme.Muted("["+defaultValue+"]"))
		}
		value, err := p.read()
		if err != nil {
			return "", err
		}
		if value == "" {
			value = defaultValue
		}
		if required && value == "" {
			fmt.Fprintln(p.out, p.theme.Error("  ! A value is required."))
			continue
		}
		return value, nil
	}
}

func (p *Prompter) Choice(label string, choices []string, defaultIndex int) (int, error) {
	if len(choices) == 0 {
		return 0, errors.New("no choices available")
	}
	for {
		PrintSection(p.out, p.theme, label)
		for i, choice := range choices {
			marker := " "
			if i == defaultIndex {
				marker = "*"
			}
			fmt.Fprintf(p.out, "  %s [%d] %s\n", p.theme.Accent(marker), i+1, choice)
		}
		fmt.Fprintf(p.out, "%s Choose a number, or press Enter for [%d]: ", p.theme.Accent(">"), defaultIndex+1)
		value, err := p.read()
		if err != nil {
			return 0, err
		}
		if value == "" {
			return defaultIndex, nil
		}
		var n int
		if _, err := fmt.Sscanf(value, "%d", &n); err != nil || n < 1 || n > len(choices) {
			fmt.Fprintln(p.out, p.theme.Error("  ! Enter a number from the displayed list."))
			continue
		}
		return n - 1, nil
	}
}

func (p *Prompter) MultiChoice(label string, choices []string) ([]int, error) {
	if len(choices) == 0 {
		return nil, errors.New("no choices available")
	}
	for {
		PrintSection(p.out, p.theme, label)
		for i, choice := range choices {
			fmt.Fprintf(p.out, "  [%d] %s\n", i+1, choice)
		}
		fmt.Fprint(p.out, p.theme.Accent("> ")+"Choose comma-separated numbers, or press Enter for none: ")
		value, err := p.read()
		if err != nil {
			return nil, err
		}
		if value == "" {
			return nil, nil
		}
		seen := make(map[int]bool)
		selected := make([]int, 0)
		valid := true
		for _, part := range strings.Split(value, ",") {
			var n int
			if _, err := fmt.Sscanf(strings.TrimSpace(part), "%d", &n); err != nil || n < 1 || n > len(choices) || seen[n] {
				valid = false
				break
			}
			seen[n] = true
			selected = append(selected, n-1)
		}
		if !valid {
			fmt.Fprintln(p.out, p.theme.Error("  ! Enter unique numbers from the displayed list, separated by commas."))
			continue
		}
		return selected, nil
	}
}

func (p *Prompter) Confirm(label string) (bool, error) {
	fmt.Fprintf(p.out, "%s %s %s: ", p.theme.Accent("?"), label, p.theme.Muted("[y/N]"))
	value, err := p.read()
	if err != nil {
		return false, err
	}
	return value == "y" || value == "yes", nil
}

func (p *Prompter) read() (string, error) {
	value, err := p.in.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	value = strings.TrimSpace(value)
	if value == "q" || value == "quit" {
		return "", ErrCanceled
	}
	if errors.Is(err, io.EOF) {
		if value != "" {
			return value, nil
		}
		return "", ErrCanceled
	}
	return value, nil
}

func IsInteractive(in io.Reader) bool {
	file, isFile := in.(osFile)
	if !isFile {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// osFile is satisfied by *os.File without importing terminal-specific packages.
type osFile interface {
	io.Reader
	Fd() uintptr
	Stat() (os.FileInfo, error)
}
