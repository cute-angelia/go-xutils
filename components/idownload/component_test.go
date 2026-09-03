package idownload

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSingleDownloadRemovesPartialFileOnCopyError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1024")
		_, _ = w.Write([]byte("partial"))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		if hijacker, ok := w.(http.Hijacker); ok {
			conn, _, err := hijacker.Hijack()
			if err == nil {
				_ = conn.Close()
			}
		}
	}))
	defer server.Close()

	dest := filepath.Join(t.TempDir(), "download.mp4")
	_, err := New(WithRetryAttempt(0)).singleDownload(context.Background(), server.URL, dest)
	if err == nil {
		t.Fatal("expected partial response to fail")
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Fatalf("expected partial file to be removed, stat err=%v", statErr)
	}
}

func TestMergeRemovesTargetFileOnError(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "download.mp4")

	err := New(WithConcurrency(1)).merge(dest)
	if err == nil {
		t.Fatal("expected merge to fail")
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Fatalf("expected target file to be removed, stat err=%v", statErr)
	}
}
