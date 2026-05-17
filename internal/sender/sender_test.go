package sender

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadURLs(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "api.txt")
	os.WriteFile(path, []byte("https://example.com/1\nhttps://example.com/2\n\n  \nhttps://example.com/3\n"), 0644)

	urls, err := readURLs(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 3 {
		t.Fatalf("expected 3 urls, got %d: %v", len(urls), urls)
	}
}
