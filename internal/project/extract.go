package project

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	MaxExtractedBytes int64 = 250 << 20
	MaxArchiveEntries       = 20000
)

func ExtractZIP(zipPath, outputParent, artifactID string) (string, error) {
	if artifactID == "" {
		return "", fmt.Errorf("project artifact ID is required")
	}
	parent, err := filepath.Abs(outputParent)
	if err != nil {
		return "", fmt.Errorf("resolving output directory: %w", err)
	}
	if err := os.MkdirAll(parent, 0755); err != nil {
		return "", fmt.Errorf("creating output directory: %w", err)
	}
	destination := filepath.Join(parent, artifactID)
	if _, err := os.Lstat(destination); err == nil {
		return "", fmt.Errorf("destination already exists: %s", destination)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("checking destination: %w", err)
	}

	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("opening archive: %w", err)
	}
	defer archive.Close()
	if len(archive.File) == 0 {
		return "", fmt.Errorf("archive is empty")
	}
	if len(archive.File) > MaxArchiveEntries {
		return "", fmt.Errorf("archive has more than %d entries", MaxArchiveEntries)
	}
	stage, err := os.MkdirTemp(parent, "."+artifactID+"-staging-")
	if err != nil {
		return "", fmt.Errorf("creating staging directory: %w", err)
	}
	completed := false
	defer func() {
		if !completed {
			_ = os.RemoveAll(stage)
		}
	}()

	seen := make(map[string]struct{}, len(archive.File))
	var total int64
	for _, entry := range archive.File {
		relative, err := archivePath(entry.Name, artifactID)
		if err != nil {
			return "", err
		}
		if relative == "" {
			continue
		}
		key := strings.ToLower(filepath.Clean(relative))
		if _, duplicate := seen[key]; duplicate {
			return "", fmt.Errorf("archive contains duplicate path %q", entry.Name)
		}
		seen[key] = struct{}{}
		mode := entry.FileInfo().Mode()
		if entry.UncompressedSize64 > uint64(MaxExtractedBytes) || total > MaxExtractedBytes-int64(entry.UncompressedSize64) {
			return "", fmt.Errorf("archive exceeds the %d MiB extraction limit", MaxExtractedBytes>>20)
		}
		total += int64(entry.UncompressedSize64)
		target := filepath.Join(stage, relative)
		if !inside(stage, target) {
			return "", fmt.Errorf("archive entry %q escapes the extraction directory", entry.Name)
		}
		if entry.FileInfo().IsDir() || strings.HasSuffix(entry.Name, "/") {
			if err := os.MkdirAll(target, 0755); err != nil {
				return "", err
			}
			continue
		}
		if mode&os.ModeSymlink != 0 || mode&os.ModeType != 0 {
			return "", fmt.Errorf("archive entry %q has an unsupported file type", entry.Name)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return "", err
		}
		if err := extractFile(entry, target, entry.UncompressedSize64); err != nil {
			return "", err
		}
	}
	if _, err := os.Lstat(destination); err == nil {
		return "", fmt.Errorf("destination already exists: %s", destination)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("checking destination: %w", err)
	}
	if err := os.Rename(stage, destination); err != nil {
		return "", fmt.Errorf("moving completed project into place: %w", err)
	}
	completed = true
	return destination, nil
}

func archivePath(name, root string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("archive contains an empty path")
	}
	cleaned := strings.ReplaceAll(name, "\\", "/")
	if strings.HasPrefix(cleaned, "/") || strings.HasPrefix(cleaned, "//") || (len(cleaned) >= 2 && cleaned[1] == ':') {
		return "", fmt.Errorf("archive path %q is absolute", name)
	}
	parts := strings.Split(cleaned, "/")
	for _, part := range parts {
		if part == ".." {
			return "", fmt.Errorf("archive path %q contains parent traversal", name)
		}
	}
	if len(parts) > 0 && parts[0] == root {
		parts = parts[1:]
	}
	if len(parts) == 0 || (len(parts) == 1 && parts[0] == "") {
		return "", nil
	}
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		if strings.Contains(part, ":") {
			return "", fmt.Errorf("archive path %q contains a drive path", name)
		}
	}
	return filepath.FromSlash(strings.Join(parts, "/")), nil
}

func inside(root, target string) bool {
	relative, err := filepath.Rel(root, target)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
}

func extractFile(entry *zip.File, target string, expected uint64) error {
	input, err := entry.Open()
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	written, copyErr := io.Copy(output, io.LimitReader(input, int64(expected)+1))
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if written != int64(expected) {
		return fmt.Errorf("archive entry %q is truncated", entry.Name)
	}
	return nil
}
