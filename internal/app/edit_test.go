package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindAndUpdateHost(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	content := "Host first\n    HostName old.example.com\n    User root\n    Port 22\n    IdentityFile ~/.ssh/old\n    ProxyJump bastion\n\nHost second\n    HostName second.example.com\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	entry, err := findEditableHost(path, "first")
	if err != nil {
		t.Fatal(err)
	}
	if entry.host.address != "old.example.com" || entry.host.user != "root" || entry.host.method != loginPrivateKey {
		t.Fatalf("unexpected parsed host: %#v", entry.host)
	}
	updated := hostConfig{name: "production", address: "new.example.com", user: "deploy", port: "2222", method: loginPassword}
	if err := updateHost(entry, updated, []string{"first", "second"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, expected := range []string{"Host production\n", "    HostName new.example.com\n", "    User deploy\n", "    Port 2222\n", "    ProxyJump bastion\n", "Host second\n"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("updated config does not contain %q:\n%s", expected, text)
		}
	}
	if strings.Contains(text, "IdentityFile") {
		t.Fatalf("old authentication directive was retained:\n%s", text)
	}
}

func TestFindEditableHostInIncludedConfig(t *testing.T) {
	dir := t.TempDir()
	included := filepath.Join(dir, "hosts.conf")
	if err := os.WriteFile(included, []byte("Host remote\n    HostName remote.example.com\n    User ubuntu\n"), 0600); err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(dir, "config")
	if err := os.WriteFile(main, []byte("Include hosts.conf\n"), 0600); err != nil {
		t.Fatal(err)
	}
	entry, err := findEditableHost(main, "remote")
	if err != nil {
		t.Fatal(err)
	}
	if entry.path != included {
		t.Fatalf("got source %q, want %q", entry.path, included)
	}
}

func TestFindEditableHostRejectsSharedBlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(path, []byte("Host one two\n    User root\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := findEditableHost(path, "one"); err == nil {
		t.Fatal("expected shared block to be rejected")
	}
}

func TestUpdateHostRejectsConcurrentChange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(path, []byte("Host server\n    HostName old.example.com\n    User root\n"), 0600); err != nil {
		t.Fatal(err)
	}
	entry, err := findEditableHost(path, "server")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("Host server\n    HostName changed.example.com\n    User root\n"), 0600); err != nil {
		t.Fatal(err)
	}
	updated := hostConfig{name: "server", address: "new.example.com", user: "root", port: "22"}
	if err := updateHost(entry, updated, []string{"server"}); err == nil {
		t.Fatal("expected concurrent change to be rejected")
	}
}
