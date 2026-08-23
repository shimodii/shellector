package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Run starts shellector with the supplied command-line arguments.
func Run(args []string) error {
	path := filepath.Join(userHomeDir(), ".ssh", "config")
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			fmt.Println("Usage: shellector [-f SSH_CONFIG]\n\nBrowse and search SSH hosts, then connect with Enter.")
			return nil
		case "-f", "--file":
			if i+1 == len(args) {
				return errors.New("-f requires a path")
			}
			i++
			path = args[i]
		default:
			return fmt.Errorf("unknown option %q", args[i])
		}
	}

	hosts, err := parseSSHConfig(path)
	if err != nil {
		return err
	}
	if len(hosts) == 0 {
		return fmt.Errorf("no named hosts found in %s", path)
	}
	if !isTerminal(os.Stdin.Fd()) || !isTerminal(os.Stdout.Fd()) {
		return errors.New("an interactive terminal is required")
	}

	host, ok, err := selectHost(hosts)
	if err != nil || !ok {
		return err
	}
	ssh, err := findExecutable("ssh")
	if err != nil {
		return err
	}
	process, err := os.StartProcess(ssh, []string{"ssh", host}, &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	})
	if err != nil {
		return fmt.Errorf("start ssh: %w", err)
	}
	state, err := process.Wait()
	if err != nil {
		return fmt.Errorf("wait for ssh: %w", err)
	}
	if !state.Success() {
		return fmt.Errorf("ssh exited with %s", state.String())
	}
	return nil
}

func findExecutable(name string) (string, error) {
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			dir = "."
		}
		candidate := filepath.Join(dir, name)
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() && info.Mode().Perm()&0111 != 0 {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("%s not found in PATH", name)
}

func userHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func parseSSHConfig(path string) ([]string, error) {
	seenFiles := make(map[string]bool)
	hostSet := make(map[string]bool)
	if err := parseConfigFile(path, seenFiles, hostSet, true); err != nil {
		return nil, err
	}
	hosts := make([]string, 0, len(hostSet))
	for host := range hostSet {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)
	return hosts, nil
}

func parseConfigFile(path string, seenFiles, hosts map[string]bool, required bool) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if seenFiles[abs] {
		return nil
	}
	seenFiles[abs] = true
	data, err := os.ReadFile(abs)
	if err != nil {
		if !required && os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read SSH config: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch strings.ToLower(fields[0]) {
		case "host":
			for _, host := range fields[1:] {
				if host != "*" && !strings.HasPrefix(host, "!") && !strings.ContainsAny(host, "*?") {
					hosts[host] = true
				}
			}
		case "include":
			for _, pattern := range fields[1:] {
				pattern = expandHome(pattern)
				if !filepath.IsAbs(pattern) {
					pattern = filepath.Join(filepath.Dir(abs), pattern)
				}
				matches, _ := filepath.Glob(pattern)
				for _, match := range matches {
					if err := parseConfigFile(match, seenFiles, hosts, false); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func expandHome(path string) string {
	if path == "~" {
		return userHomeDir()
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(userHomeDir(), path[2:])
	}
	return path
}
