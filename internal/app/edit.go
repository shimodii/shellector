package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type editableHost struct {
	host            hostConfig
	originalName    string
	path            string
	start, end      int
	originalBlock   []byte
	extraDirectives []string
}

func findEditableHost(path, name string) (editableHost, error) {
	entry, found, err := findEditableHostFile(path, name, make(map[string]bool), true)
	if err != nil {
		return editableHost{}, err
	}
	if !found {
		return editableHost{}, fmt.Errorf("host %q could not be located in SSH config", name)
	}
	return entry, nil
}

func findEditableHostFile(path, name string, seen map[string]bool, required bool) (editableHost, bool, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return editableHost{}, false, err
	}
	if seen[abs] {
		return editableHost{}, false, nil
	}
	seen[abs] = true
	data, err := os.ReadFile(abs)
	if err != nil {
		if !required && os.IsNotExist(err) {
			return editableHost{}, false, nil
		}
		return editableHost{}, false, fmt.Errorf("read SSH config: %w", err)
	}

	lines := splitLines(data)
	for i, line := range lines {
		fields := configFields(line.text)
		if len(fields) < 2 {
			continue
		}
		switch strings.ToLower(fields[0]) {
		case "include":
			for _, pattern := range fields[1:] {
				pattern = expandHome(pattern)
				if !filepath.IsAbs(pattern) {
					pattern = filepath.Join(filepath.Dir(abs), pattern)
				}
				matches, _ := filepath.Glob(pattern)
				for _, match := range matches {
					entry, found, err := findEditableHostFile(match, name, seen, false)
					if err != nil || found {
						return entry, found, err
					}
				}
			}
		case "host":
			if !contains(fields[1:], name) {
				continue
			}
			if len(fields) != 2 {
				return editableHost{}, false, fmt.Errorf("host %q shares a Host block with other aliases and cannot be edited safely", name)
			}
			endLine := i + 1
			for endLine < len(lines) {
				next := configFields(lines[endLine].text)
				if len(next) > 0 && (strings.EqualFold(next[0], "Host") || strings.EqualFold(next[0], "Match")) {
					break
				}
				endLine++
			}
			entry := editableHost{
				host:         hostConfig{name: name, address: name, port: "22"},
				originalName: name,
				path:         abs,
				start:        line.start,
				end:          len(data),
			}
			if endLine < len(lines) {
				entry.end = lines[endLine].start
			}
			entry.originalBlock = append([]byte(nil), data[entry.start:entry.end]...)
			parseEditableDirectives(&entry, lines[i+1:endLine])
			return entry, true, nil
		}
	}
	return editableHost{}, false, nil
}

type configLine struct {
	text       string
	start, end int
}

func splitLines(data []byte) []configLine {
	parts := strings.SplitAfter(string(data), "\n")
	lines := make([]configLine, 0, len(parts))
	offset := 0
	for _, part := range parts {
		if part == "" {
			continue
		}
		lines = append(lines, configLine{text: part, start: offset, end: offset + len(part)})
		offset += len(part)
	}
	return lines
}

func configFields(line string) []string {
	return strings.Fields(strings.TrimSpace(strings.SplitN(line, "#", 2)[0]))
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func parseEditableDirectives(entry *editableHost, lines []configLine) {
	for _, line := range lines {
		fields := configFields(line.text)
		if len(fields) < 2 {
			entry.extraDirectives = append(entry.extraDirectives, line.text)
			continue
		}
		switch strings.ToLower(fields[0]) {
		case "hostname":
			entry.host.address = fields[1]
		case "user":
			entry.host.user = fields[1]
		case "port":
			if _, err := strconv.Atoi(fields[1]); err == nil {
				entry.host.port = fields[1]
			}
		case "identityfile":
			entry.host.privateKey = fields[1]
			entry.host.method = loginPrivateKey
		case "preferredauthentications":
			if strings.Contains(strings.ToLower(fields[1]), "publickey") {
				entry.host.method = loginPrivateKey
			}
		case "identitiesonly":
		default:
			entry.extraDirectives = append(entry.extraDirectives, line.text)
		}
	}
}

func updateHost(entry editableHost, updated hostConfig, existing []string) error {
	if err := validateHost(updated); err != nil {
		return err
	}
	if updated.name != entry.originalName && contains(existing, updated.name) {
		return fmt.Errorf("host %q already exists", updated.name)
	}
	data, err := os.ReadFile(entry.path)
	if err != nil {
		return fmt.Errorf("read SSH config: %w", err)
	}
	if entry.start < 0 || entry.end < entry.start || entry.end > len(data) {
		return errors.New("SSH config changed while editing; reopen the host and try again")
	}
	if string(data[entry.start:entry.end]) != string(entry.originalBlock) {
		return errors.New("SSH config changed while editing; reopen the host and try again")
	}
	replacement := formatHostBlock(updated, entry.extraDirectives)
	result := append([]byte{}, data[:entry.start]...)
	result = append(result, replacement...)
	result = append(result, data[entry.end:]...)

	info, err := os.Stat(entry.path)
	if err != nil {
		return fmt.Errorf("stat SSH config: %w", err)
	}
	temp, err := os.CreateTemp(filepath.Dir(entry.path), ".shellector-config-*")
	if err != nil {
		return fmt.Errorf("create temporary SSH config: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(info.Mode().Perm()); err != nil {
		temp.Close()
		return fmt.Errorf("set SSH config permissions: %w", err)
	}
	if _, err := temp.Write(result); err != nil {
		temp.Close()
		return fmt.Errorf("write SSH config: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close SSH config: %w", err)
	}
	if err := os.Rename(tempPath, entry.path); err != nil {
		return fmt.Errorf("replace SSH config: %w", err)
	}
	return nil
}
