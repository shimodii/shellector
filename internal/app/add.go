package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

type loginMethod int

const (
	loginPassword loginMethod = iota
	loginPrivateKey
)

type hostConfig struct {
	name       string
	address    string
	user       string
	port       string
	method     loginMethod
	privateKey string
}

func promptAddHost() (hostConfig, bool, error) {
	return promptHost("Add SSH host", hostConfig{port: "22"})
}

func promptEditHost(host hostConfig) (hostConfig, bool, error) {
	return promptHost("Edit SSH host", host)
}

func promptHost(title string, host hostConfig) (hostConfig, bool, error) {
	state, err := makeRaw(os.Stdin.Fd())
	if err != nil {
		return hostConfig{}, false, err
	}
	fmt.Print("\x1b[?1049h\x1b[?25l")
	defer func() {
		fmt.Print("\x1b[?25h\x1b[?1049l")
		restoreTerminal(os.Stdin.Fd(), state)
	}()

	field, message := 0, ""
	for {
		renderHostForm(title, host, field, message)
		key, text, err := readKey()
		if err != nil {
			return hostConfig{}, false, err
		}
		message = ""
		switch key {
		case "escape", "ctrl-c":
			return hostConfig{}, false, nil
		case "up", "down":
			if field == 4 {
				host.method = loginMethod(1 - int(host.method))
			}
		case "tab":
			field = nextField(field, host.method)
		case "shift-tab":
			field = previousField(field, host.method)
		case "enter":
			if field < lastField(host.method) {
				field = nextField(field, host.method)
				continue
			}
			if err := validateHost(host); err != nil {
				message = err.Error()
				continue
			}
			return host, true, nil
		case "backspace":
			value := editableField(&host, field)
			if value != nil && *value != "" {
				_, n := utf8.DecodeLastRuneInString(*value)
				*value = (*value)[:len(*value)-n]
			}
		case "text":
			if value := editableField(&host, field); value != nil {
				*value += text
			}
		}
	}
}

func nextField(field int, method loginMethod) int {
	if field >= lastField(method) {
		return 0
	}
	return field + 1
}

func previousField(field int, method loginMethod) int {
	if field <= 0 {
		return lastField(method)
	}
	return field - 1
}

func lastField(method loginMethod) int {
	if method == loginPrivateKey {
		return 5
	}
	return 4
}

func editableField(host *hostConfig, field int) *string {
	switch field {
	case 0:
		return &host.name
	case 1:
		return &host.address
	case 2:
		return &host.user
	case 3:
		return &host.port
	case 5:
		return &host.privateKey
	default:
		return nil
	}
}

func renderAddHost(host hostConfig, field int, message string) {
	renderHostForm("Add SSH host", host, field, message)
}

func formatAddHost(host hostConfig, field int, message string) string {
	return formatHostForm("Add SSH host", host, field, message)
}

func renderHostForm(title string, host hostConfig, field int, message string) {
	fmt.Print(formatHostForm(title, host, field, message))
}

func formatHostForm(title string, host hostConfig, field int, message string) string {
	fields := []struct{ label, value string }{
		{"Name", host.name},
		{"Address", host.address},
		{"User", host.user},
		{"Port", host.port},
	}
	if host.method == loginPrivateKey {
		fields = append(fields, struct{ label, value string }{"Private key path", host.privateKey})
	}

	var b strings.Builder
	fmt.Fprintf(&b, "\x1b[H\x1b[2J\x1b[1;36m  %s\x1b[0m\n\n", title)
	for i, input := range fields[:4] {
		if i == field {
			fmt.Fprintf(&b, "  \x1b[1;36m› %s\x1b[0m\n", input.label)
		} else {
			fmt.Fprintf(&b, "    %s\n", input.label)
		}
		fmt.Fprintf(&b, "      %s", input.value)
		if i == field {
			b.WriteString("\x1b[7m \x1b[0m")
		}
		b.WriteString("\n\n")
	}

	if field == 4 {
		b.WriteString("  \x1b[1;36m› Login method\x1b[0m\n")
	} else {
		b.WriteString("    Login method\n")
	}
	if host.method == loginPassword {
		b.WriteString("      \x1b[1;36m● Password\x1b[0m    ○ Private key\n\n")
	} else {
		b.WriteString("      ○ Password    \x1b[1;36m● Private key\x1b[0m\n\n")
	}

	if host.method == loginPrivateKey {
		input := fields[4]
		if field == 5 {
			fmt.Fprintf(&b, "  \x1b[1;36m› %s\x1b[0m\n", input.label)
		} else {
			fmt.Fprintf(&b, "    %s\n", input.label)
		}
		fmt.Fprintf(&b, "      %s", input.value)
		if field == 5 {
			b.WriteString("\x1b[7m \x1b[0m")
		}
		b.WriteString("\n\n")
	}
	if message != "" {
		fmt.Fprintf(&b, "  \x1b[31m%s\x1b[0m\n\n", message)
	}
	b.WriteString("  \x1b[2mtab/enter next  shift+tab back  ↑/↓ choose method  enter to save  esc cancel\x1b[0m")
	return b.String()
}

func validateHost(host hostConfig) error {
	if host.name == "" || strings.ContainsAny(host.name, " \t\r\n*?!#") {
		return errors.New("name is required and cannot contain spaces or wildcard characters")
	}
	if host.address == "" || strings.ContainsAny(host.address, " \t\r\n#") {
		return errors.New("address is required and cannot contain spaces")
	}
	if host.user == "" || strings.ContainsAny(host.user, " \t\r\n#") {
		return errors.New("user is required and cannot contain spaces")
	}
	port, err := strconv.Atoi(host.port)
	if err != nil || port < 1 || port > 65535 {
		return errors.New("port must be a number between 1 and 65535")
	}
	if host.method == loginPrivateKey && (host.privateKey == "" || strings.ContainsAny(host.privateKey, " \t\r\n#")) {
		return errors.New("private key path is required and cannot contain spaces")
	}
	return nil
}

func appendHost(path string, host hostConfig, existing []string) error {
	if err := validateHost(host); err != nil {
		return err
	}
	for _, name := range existing {
		if name == host.name {
			return fmt.Errorf("host %q already exists", host.name)
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create SSH directory: %w", err)
	}

	separator := ""
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		separator = "\n"
		if data[len(data)-1] != '\n' {
			separator = "\n\n"
		}
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read SSH config: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("open SSH config: %w", err)
	}
	defer file.Close()
	if _, err := file.WriteString(separator + formatHostBlock(host, nil)); err != nil {
		return fmt.Errorf("write SSH config: %w", err)
	}
	return nil
}

func formatHostBlock(host hostConfig, extra []string) string {
	var block strings.Builder
	fmt.Fprintf(&block, "Host %s\n    HostName %s\n    User %s\n    Port %s\n", host.name, host.address, host.user, host.port)
	if host.method == loginPrivateKey {
		fmt.Fprintf(&block, "    IdentityFile %s\n    IdentitiesOnly yes\n    PreferredAuthentications publickey\n", host.privateKey)
	} else {
		block.WriteString("    PreferredAuthentications password,keyboard-interactive\n")
	}
	for _, line := range extra {
		block.WriteString(line)
		if !strings.HasSuffix(line, "\n") {
			block.WriteByte('\n')
		}
	}
	return block.String()
}
