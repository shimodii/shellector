package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeleteHostRemovesOnlySelectedBlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	content := "Host first\n    HostName first.example.com\n    User root\n\nHost second\n    HostName second.example.com\n    User ubuntu\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	entry, err := findEditableHost(path, "first")
	if err != nil {
		t.Fatal(err)
	}
	if err := deleteHost(entry); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "Host first") {
		t.Fatalf("deleted host remains in config:\n%s", data)
	}
	if !strings.Contains(string(data), "Host second") {
		t.Fatalf("neighboring host was removed:\n%s", data)
	}
}

func TestDeletePromptNamesHostAndRequiresConfirmation(t *testing.T) {
	prompt := formatDeletePrompt("production")
	if !strings.Contains(prompt, "production") || !strings.Contains(prompt, "Y confirm") {
		t.Fatalf("unexpected delete prompt: %q", prompt)
	}
}
