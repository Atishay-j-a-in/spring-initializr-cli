package project

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func writeArchive(t *testing.T, entries map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "project.zip")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for name, content := range entries {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExtractZIP(t *testing.T) {
	t.Run("valid archive", func(t *testing.T) {
		parent := t.TempDir()
		path := writeArchive(t, map[string]string{"demo/README.md": "hello"})
		destination, err := ExtractZIP(path, parent, "demo")
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(destination, "README.md"))
		if err != nil || string(data) != "hello" {
			t.Fatalf("%q %v", data, err)
		}
	})
	for _, name := range []string{"../escape.txt", "/absolute.txt", "C:/drive.txt"} {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			path := writeArchive(t, map[string]string{name: "bad"})
			if _, err := ExtractZIP(path, parent, "demo"); err == nil {
				t.Fatal("expected error")
			}
			if _, err := os.Stat(filepath.Join(parent, "demo")); !os.IsNotExist(err) {
				t.Fatal("partial destination exists")
			}
		})
	}
	t.Run("existing target", func(t *testing.T) {
		parent := t.TempDir()
		if err := os.Mkdir(filepath.Join(parent, "demo"), 0755); err != nil {
			t.Fatal(err)
		}
		path := writeArchive(t, map[string]string{"demo/a": "x"})
		if _, err := ExtractZIP(path, parent, "demo"); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("symlink", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "project.zip")
		file, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		writer := zip.NewWriter(file)
		header := &zip.FileHeader{Name: "demo/link"}
		header.SetMode(os.ModeSymlink | 0777)
		entry, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte("target")); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := ExtractZIP(path, t.TempDir(), "demo"); err == nil {
			t.Fatal("expected symlink rejection")
		}
	})
}
