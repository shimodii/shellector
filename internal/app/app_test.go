package app

import (
	"os"
	"path/filepath"
	"reflect"
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
