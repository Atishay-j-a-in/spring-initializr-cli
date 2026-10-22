package cli

import (
	"bytes"
	"strings"
	"testing"

	"example.com/spring-create/internal/initializr"
)

func wizardMetadata() initializr.Metadata {
	return initializr.Metadata{
		GroupID:     initializr.TextCapability{Default: "com.example"},
		ArtifactID:  initializr.TextCapability{Default: "demo"},
		PackageName: initializr.TextCapability{Default: "com.example.demo"},
		Type: initializr.SelectCapability{Default: "maven-project", Values: []initializr.Value{
			{ID: "maven-project", Name: "Maven", Action: "/starter.zip", Tags: map[string]string{"format": "project"}},
		}},
		Language:    initializr.SelectCapability{Default: "java", Values: []initializr.Value{{ID: "java", Name: "Java"}}},
		Packaging:   initializr.SelectCapability{Default: "jar", Values: []initializr.Value{{ID: "jar", Name: "Jar"}}},
		JavaVersion: initializr.SelectCapability{Default: "17", Values: []initializr.Value{{ID: "17", Name: "17"}}},
		BootVersion: initializr.SelectCapability{Default: "4", Values: []initializr.Value{{ID: "4", Name: "4"}}},
		Dependencies: initializr.DependencyCapability{Values: []initializr.DependencyGroup{
			{Name: "Web", Values: []initializr.Dependency{{ID: "web", Name: "Web"}}},
		}},
	}
}

func TestRunWizard(t *testing.T) {
	var out bytes.Buffer
	input := strings.NewReader("demo\n" + strings.Repeat("\n", 9) + "1\n1\n\ny\n")
	config, err := RunWizard(input, &out, wizardMetadata(), t.TempDir(), Theme{})
	if err != nil {
		t.Fatal(err)
	}
	if config.ArtifactID != "demo" || len(config.Dependencies) != 1 || config.Dependencies[0] != "web" {
		t.Fatalf("unexpected config: %#v", config)
	}
}

func TestRunWizardCancellation(t *testing.T) {
	_, err := RunWizard(strings.NewReader("q\n"), &bytes.Buffer{}, wizardMetadata(), t.TempDir(), Theme{})
	if err != ErrCanceled {
		t.Fatalf("expected cancellation, got %v", err)
	}
}
