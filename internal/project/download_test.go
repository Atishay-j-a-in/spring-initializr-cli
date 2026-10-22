package project

import (
	"archive/zip"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func testZIP(t *testing.T) []byte {
	t.Helper()
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	f, _ := z.Create("demo/file.txt")
	f.Write([]byte("ok"))
	z.Close()
	return b.Bytes()
}
func TestDownloadZIP(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write(testZIP(t)) }))
		defer s.Close()
		path, err := DownloadZIP(context.Background(), s.Client(), s.URL, "test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(path)
		if _, err := os.Stat(path); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("status", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "bad", 400) }))
		defer s.Close()
		if _, err := DownloadZIP(context.Background(), s.Client(), s.URL, "test"); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("invalid zip", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("<html>error</html>")) }))
		defer s.Close()
		if _, err := DownloadZIP(context.Background(), s.Client(), s.URL, "test"); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("timeout", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(100 * time.Millisecond) }))
		defer s.Close()
		client := &http.Client{Timeout: time.Millisecond}
		if _, err := DownloadZIP(context.Background(), client, s.URL, "test"); err == nil {
			t.Fatal("expected error")
		}
	})
}
