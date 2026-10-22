package initializr

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const fixture = `{"dependencies":{"type":"hierarchical-multi-select","values":[{"name":"Web","values":[{"id":"web","name":"Spring Web"}]}]},"type":{"type":"action","default":"maven-project","values":[{"id":"maven-project","name":"Maven","action":"/starter.zip","tags":{"build":"maven","format":"project"}}]},"packaging":{"type":"single-select","default":"jar","values":[{"id":"jar","name":"Jar"}]},"javaVersion":{"type":"single-select","default":"17","values":[{"id":"17","name":"17"}]},"language":{"type":"single-select","default":"java","values":[{"id":"java","name":"Java"}]},"bootVersion":{"type":"single-select","default":"4.1.1","values":[{"id":"4.1.1","name":"4.1.1"}]},"groupId":{"type":"text","default":"com.example"},"artifactId":{"type":"text","default":"demo"},"name":{"type":"text"},"description":{"type":"text"},"packageName":{"type":"text","default":"com.example.demo"}}`

func TestFetchMetadata(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Accept") != MediaType {
				t.Error("accept header")
			}
			w.Write([]byte(fixture))
		}))
		defer s.Close()
		client, _ := NewClient(s.URL, "test", time.Second)
		metadata, err := client.FetchMetadata(context.Background())
		if err != nil || metadata.Type.Default != "maven-project" {
			t.Fatalf("%v %#v", err, metadata)
		}
	})
	t.Run("bad status", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
		}))
		defer s.Close()
		client, _ := NewClient(s.URL, "test", time.Second)
		if _, err := client.FetchMetadata(context.Background()); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("invalid JSON", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("{")) }))
		defer s.Close()
		client, _ := NewClient(s.URL, "test", time.Second)
		if _, err := client.FetchMetadata(context.Background()); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("timeout", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(100 * time.Millisecond) }))
		defer s.Close()
		client, _ := NewClient(s.URL, "test", time.Millisecond)
		if _, err := client.FetchMetadata(context.Background()); err == nil {
			t.Fatal("expected error")
		}
	})
}
