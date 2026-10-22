package initializr

import (
	"net/url"
	"testing"
)

func TestGenerationURL(t *testing.T) {
	client, err := NewClient("https://example.test/base", "test", 0)
	if err != nil {
		t.Fatal(err)
	}
	requestURL := client.GenerationURL(GenerationRequest{Type: "maven-project", Language: "java", BaseDir: "my app", GroupID: "com.example", ArtifactID: "my-app", Name: "My App", Description: "A & B", PackageName: "com.example.app", Packaging: "jar", JavaVersion: "17", Dependencies: []string{"web", "data-jpa"}})
	u, err := url.Parse(requestURL)
	if err != nil {
		t.Fatal(err)
	}
	if u.Path != "/base/starter.zip" {
		t.Fatalf("unexpected path: %s", u.Path)
	}
	query := u.Query()
	if query.Get("name") != "My App" || query.Get("description") != "A & B" || query.Get("dependencies") != "web,data-jpa" {
		t.Fatalf("unexpected query: %s", u.RawQuery)
	}
}
