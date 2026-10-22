package project

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	MaxDownloadBytes  int64 = 100 << 20
	GenerationTimeout       = 60 * time.Second
)

func DownloadZIP(ctx context.Context, client *http.Client, requestURL, userAgent string) (string, error) {
	if client == nil {
		client = &http.Client{Timeout: GenerationTimeout}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/zip")
	req.Header.Set("User-Agent", userAgent)
	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("downloading generated project: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 8192))
		return "", fmt.Errorf("project generation returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	file, err := os.CreateTemp("", "spring-create-*.zip")
	if err != nil {
		return "", err
	}
	path := file.Name()
	defer func() {
		if err != nil {
			os.Remove(path)
		}
	}()
	written, copyErr := io.Copy(file, io.LimitReader(response.Body, MaxDownloadBytes+1))
	closeErr := file.Close()
	if copyErr != nil {
		err = copyErr
		return "", err
	}
	if closeErr != nil {
		err = closeErr
		return "", err
	}
	if written > MaxDownloadBytes {
		err = fmt.Errorf("generated archive exceeds the %d MiB download limit", MaxDownloadBytes>>20)
		return "", err
	}
	archive, openErr := zip.OpenReader(path)
	if openErr != nil {
		err = fmt.Errorf("generated response is not a valid ZIP archive: %w", openErr)
		return "", err
	}
	if closeErr := archive.Close(); closeErr != nil {
		err = closeErr
		return "", err
	}
	return path, nil
}
