package app_test

import (
	"archive/zip"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"example.com/spring-create/internal/cli"
	"example.com/spring-create/internal/initializr"
	"example.com/spring-create/internal/project"
)

const metadataFixture = `{"dependencies":{"values":[{"name":"Web","values":[{"id":"web","name":"Web"}]}]},"type":{"default":"maven-project","values":[{"id":"maven-project","name":"Maven","action":"/starter.zip","tags":{"format":"project"}}]},"packaging":{"default":"jar","values":[{"id":"jar","name":"Jar"}]},"javaVersion":{"default":"17","values":[{"id":"17","name":"17"}]},"language":{"default":"java","values":[{"id":"java","name":"Java"}]},"bootVersion":{"default":"4","values":[{"id":"4","name":"4"}]},"groupId":{"default":"com.example"},"artifactId":{"default":"demo"},"name":{},"description":{},"packageName":{"default":"com.example.demo"}}`

func TestFakeInitializrFlow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Write([]byte(metadataFixture))
			return
		}
		if r.URL.Path != "/starter.zip" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("dependencies") != "web" || r.URL.Query().Get("baseDir") != "demo" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		writer := zip.NewWriter(w)
		file, _ := writer.Create("demo/README.md")
		file.Write([]byte("generated"))
		writer.Close()
	}))
	defer server.Close()
	client, err := initializr.NewClient(server.URL, "test", 0)
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := client.FetchMetadata(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	config, err := cli.RunWizard(strings.NewReader("demo\n"+strings.Repeat("\n", 9)+"1\n1\n\ny\n"), &output, metadata, t.TempDir(), cli.Theme{})
	if err != nil {
		t.Fatal(err)
	}
	archive, err := project.DownloadZIP(context.Background(), server.Client(), client.GenerationURL(initializr.GenerationRequest{Type: config.Type, Language: config.Language, BootVersion: config.BootVersion, BaseDir: config.ArtifactID, GroupID: config.GroupID, ArtifactID: config.ArtifactID, Name: config.Name, PackageName: config.PackageName, Packaging: config.Packaging, JavaVersion: config.JavaVersion, Dependencies: config.Dependencies}), "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(archive)
	destination, err := project.ExtractZIP(archive, config.OutputParent, config.ArtifactID)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(destination, "README.md"))
	if err != nil || string(data) != "generated" {
		t.Fatalf("%q, %v", data, err)
	}
}
