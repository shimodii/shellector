package app

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseSSHConfig(t *testing.T) {
	dir := t.TempDir()
	inc := filepath.Join(dir, "work.conf")
	if err := os.WriteFile(inc, []byte("Host prod db-* !blocked\n"), 0600); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(dir, "config")
	content := "Host github personal\n  HostName github.com\nHost *\nInclude work.conf\nHost github # duplicate\n"
	if err := os.WriteFile(config, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	hosts, err := parseSSHConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"github", "personal", "prod"}
	if !reflect.DeepEqual(hosts, want) {
		t.Fatalf("got %v, want %v", hosts, want)
	}
}

func TestFilterHosts(t *testing.T) {
	hosts := []string{"dev-api", "prod-api", "prod-db"}
	got := filterHosts(hosts, "PROD api")
	want := []string{"prod-api"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAppendPasswordHost(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".ssh", "config")
	host := hostConfig{name: "staging", address: "192.0.2.10", user: "deploy", port: "2222", method: loginPassword}
	if err := appendHost(path, host, nil); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "Host staging\n    HostName 192.0.2.10\n    User deploy\n    Port 2222\n    PreferredAuthentications password,keyboard-interactive\n"
	if string(data) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", data, want)
	}
}

func TestAppendPrivateKeyHost(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(path, []byte("Host existing\n    HostName example.com\n"), 0600); err != nil {
		t.Fatal(err)
	}
	host := hostConfig{name: "prod", address: "prod.example.com", user: "ubuntu", port: "22", method: loginPrivateKey, privateKey: "~/.ssh/id_ed25519"}
	if err := appendHost(path, host, []string{"existing"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "Host existing\n    HostName example.com\n\nHost prod\n    HostName prod.example.com\n    User ubuntu\n    Port 22\n    IdentityFile ~/.ssh/id_ed25519\n    IdentitiesOnly yes\n    PreferredAuthentications publickey\n"
	if string(data) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", data, want)
	}
}

func TestAppendHostRejectsInvalidAndDuplicateHosts(t *testing.T) {
	tests := []hostConfig{
		{name: "bad host", address: "example.com", user: "ubuntu", port: "22"},
		{name: "valid", address: "example.com", user: "ubuntu", port: "70000"},
		{name: "valid", address: "example.com", user: "ubuntu", port: "22", method: loginPrivateKey},
		{name: "valid", address: "example.com", port: "22"},
	}
	for _, host := range tests {
		if err := appendHost(filepath.Join(t.TempDir(), "config"), host, nil); err == nil {
			t.Fatalf("expected validation error for %#v", host)
		}
	}
	host := hostConfig{name: "duplicate", address: "example.com", user: "ubuntu", port: "22"}
	if err := appendHost(filepath.Join(t.TempDir(), "config"), host, []string{"duplicate"}); err == nil {
		t.Fatal("expected duplicate host error")
	}
}

func TestAddHostFormUsesVerticalFieldLayout(t *testing.T) {
	output := formatAddHost(hostConfig{port: "22"}, 0, "")
	for _, field := range []string{"Name\x1b[0m\n", "Address\n", "User\n", "Port\n", "Login method\n"} {
		if !strings.Contains(output, field) {
			t.Fatalf("form does not contain vertically rendered field %q", field)
		}
	}
	if strings.Contains(output, "Name:         ") {
		t.Fatal("form still uses padded inline labels")
	}
}

func TestAddHostFormHighlightsActiveTitle(t *testing.T) {
	output := formatAddHost(hostConfig{port: "22"}, 2, "")
	if !strings.Contains(output, "\x1b[1;36m› User\x1b[0m") {
		t.Fatal("active User title is not highlighted")
	}
	if strings.Contains(output, "\x1b[1;36m› Name\x1b[0m") {
		t.Fatal("inactive Name title is highlighted")
	}
}

func TestFormFieldNavigation(t *testing.T) {
	if got := previousField(0, loginPassword); got != 4 {
		t.Fatalf("Shift+Tab from first password field = %d, want 4", got)
	}
	if got := previousField(0, loginPrivateKey); got != 5 {
		t.Fatalf("Shift+Tab from first private-key field = %d, want 5", got)
	}
	if got := previousField(3, loginPassword); got != 2 {
		t.Fatalf("Shift+Tab from port = %d, want 2", got)
	}
}
