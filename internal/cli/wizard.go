package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"example.com/spring-create/internal/initializr"
	"example.com/spring-create/internal/validate"
)

type ProjectConfig struct {
	Name         string
	GroupID      string
	ArtifactID   string
	PackageName  string
	Description  string
	Language     string
	Type         string
	Packaging    string
	JavaVersion  string
	BootVersion  string
	Dependencies []string
	OutputParent string
}

func (c ProjectConfig) Destination() string { return filepath.Join(c.OutputParent, c.ArtifactID) }

func RunWizard(in io.Reader, out io.Writer, metadata initializr.Metadata, outputDefault string, theme Theme) (ProjectConfig, error) {
	p := NewPrompter(in, out, theme)
	PrintBanner(out, theme)
	fmt.Fprintln(out, "Spring Create uses https://start.spring.io to generate your project.")
	fmt.Fprintln(out, theme.Muted("Press Enter to keep a default. Type q at any prompt to cancel."))
	PrintSection(out, theme, "1. Project identity")
	name, err := promptValidated(p, "Project name", metadata.ArtifactID.Default, true, validate.SafeName)
	if err != nil {
		return ProjectConfig{}, err
	}
	groupID, err := promptValidated(p, "Group ID", metadata.GroupID.Default, true, validate.GroupID)
	if err != nil {
		return ProjectConfig{}, err
	}
	artifact, err := promptValidated(p, "Artifact ID", name, true, validate.SafeName)
	if err != nil {
		return ProjectConfig{}, err
	}
	packageName, err := promptValidated(p, "Base package name", metadata.PackageName.Default, true, validate.JavaPackage)
	if err != nil {
		return ProjectConfig{}, err
	}
	description, err := promptValidated(p, "Description", metadata.Description.Default, false, validate.Description)
	if err != nil {
		return ProjectConfig{}, err
	}

	PrintSection(out, theme, "2. Project options")
	language, err := selectCapability(p, "Language", metadata.Language)
	if err != nil {
		return ProjectConfig{}, err
	}
	projectTypes := projectTypes(metadata.Type)
	buildType, err := selectCapability(p, "Build system", projectTypes)
	if err != nil {
		return ProjectConfig{}, err
	}
	packaging, err := selectCapability(p, "Packaging", metadata.Packaging)
	if err != nil {
		return ProjectConfig{}, err
	}
	javaVersion, err := selectCapability(p, "Java version", metadata.JavaVersion)
	if err != nil {
		return ProjectConfig{}, err
	}
	bootVersion, err := selectCapability(p, "Spring Boot version", metadata.BootVersion)
	if err != nil {
		return ProjectConfig{}, err
	}
	PrintSection(out, theme, "3. Dependencies")
	dependencies, err := chooseDependencies(p, metadata.Dependencies)
	if err != nil {
		return ProjectConfig{}, err
	}
	if outputDefault == "" {
		outputDefault, _ = filepath.Abs(".")
	}
	PrintSection(out, theme, "4. Output location")
	var outputParent string
	for {
		outputParent, err = promptValidated(p, "Output directory", outputDefault, true, func(s string) error { _, e := validate.OutputPath(s); return e })
		if err != nil {
			return ProjectConfig{}, err
		}
		outputParent, _ = validate.OutputPath(outputParent)
		if _, statErr := os.Lstat(filepath.Join(outputParent, artifact)); os.IsNotExist(statErr) {
			break
		} else if statErr == nil {
			fmt.Fprintln(out, theme.Warning("! That project destination already exists. Choose another output directory or artifact ID."))
			continue
		} else {
			fmt.Fprintln(out, theme.Error("! Cannot inspect the requested output directory:"), statErr)
		}
	}

	config := ProjectConfig{Name: name, GroupID: groupID, ArtifactID: artifact, PackageName: packageName, Description: description, Language: language, Type: buildType, Packaging: packaging, JavaVersion: javaVersion, BootVersion: bootVersion, Dependencies: dependencies, OutputParent: outputParent}
	printSummary(out, config, theme)
	confirmed, err := p.Confirm("Generate this project")
	if err != nil {
		return ProjectConfig{}, err
	}
	if !confirmed {
		return ProjectConfig{}, ErrCanceled
	}
	return config, nil
}

func promptValidated(p *Prompter, label, defaultValue string, required bool, validateValue func(string) error) (string, error) {
	for {
		value, err := p.Text(label, defaultValue, required)
		if err != nil {
			return "", err
		}
		if err := validateValue(value); err != nil {
			fmt.Fprintln(p.out, p.theme.Error("  ! "+err.Error()))
			continue
		}
		return value, nil
	}
}

func selectCapability(p *Prompter, label string, capability initializr.SelectCapability) (string, error) {
	choices := make([]string, len(capability.Values))
	defaultIndex := 0
	for i, value := range capability.Values {
		choices[i] = value.Name
		if value.ID == capability.Default {
			defaultIndex = i
		}
	}
	index, err := p.Choice(label, choices, defaultIndex)
	if err != nil {
		return "", err
	}
	return capability.Values[index].ID, nil
}

func projectTypes(capability initializr.SelectCapability) initializr.SelectCapability {
	filtered := initializr.SelectCapability{Type: capability.Type, Default: capability.Default}
	for _, value := range capability.Values {
		if value.Action == "/starter.zip" && value.Tags["format"] == "project" {
			filtered.Values = append(filtered.Values, value)
		}
	}
	if _, ok := filtered.Find(filtered.Default); !ok && len(filtered.Values) > 0 {
		filtered.Default = filtered.Values[0].ID
	}
	return filtered
}

func chooseDependencies(p *Prompter, capability initializr.DependencyCapability) ([]string, error) {
	groups := make([]string, len(capability.Values))
	for i, group := range capability.Values {
		groups[i] = group.Name
	}
	selectedGroups, err := p.MultiChoice("Dependency categories", groups)
	if err != nil {
		return nil, err
	}
	var dependencies []string
	for _, groupIndex := range selectedGroups {
		group := capability.Values[groupIndex]
		choices := make([]string, len(group.Values))
		for i, dependency := range group.Values {
			choices[i] = fmt.Sprintf("%s (%s)", dependency.Name, dependency.ID)
		}
		selected, err := p.MultiChoice(group.Name+" dependencies", choices)
		if err != nil {
			return nil, err
		}
		for _, index := range selected {
			dependencies = append(dependencies, group.Values[index].ID)
		}
	}
	return dependencies, nil
}

func printSummary(out io.Writer, c ProjectConfig, theme Theme) {
	PrintSection(out, theme, "Review configuration")
	for _, row := range [][2]string{{"Name", c.Name}, {"Group ID", c.GroupID}, {"Artifact ID", c.ArtifactID}, {"Package", c.PackageName}, {"Description", c.Description}, {"Language", c.Language}, {"Build", c.Type}, {"Packaging", c.Packaging}, {"Java", c.JavaVersion}, {"Spring Boot", c.BootVersion}, {"Destination", c.Destination()}} {
		fmt.Fprintf(out, "  %-14s %s\n", row[0]+":", row[1])
	}
	if len(c.Dependencies) == 0 {
		fmt.Fprintln(out, "  Dependencies:  none selected")
	} else {
		fmt.Fprintln(out, "  Dependencies: ", strings.Join(c.Dependencies, ", "))
	}
	fmt.Fprintln(out, theme.Muted(rule))
}
