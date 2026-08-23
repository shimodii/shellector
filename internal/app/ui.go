package app

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

func selectHost(hosts []string) (selected string, ok bool, err error) {
	state, err := makeRaw(os.Stdin.Fd())
	if err != nil {
		return "", false, err
	}
	fmt.Print("\x1b[?1049h\x1b[?25l")
	defer func() {
		fmt.Print("\x1b[?25h\x1b[?1049l")
		restoreTerminal(os.Stdin.Fd(), state)
	}()

	query, cursor := "", 0
	for {
		matches := filterHosts(hosts, query)
		if cursor >= len(matches) {
			cursor = max(0, len(matches)-1)
		}
		render(matches, query, cursor)
		key, text, readErr := readKey()
		if readErr != nil {
			return "", false, readErr
		}
		switch key {
		case "up":
			if cursor > 0 {
				cursor--
			}
		case "down":
			if cursor+1 < len(matches) {
				cursor++
			}
		case "enter":
			if len(matches) > 0 {
				return matches[cursor], true, nil
			}
		case "escape", "ctrl-c":
			return "", false, nil
		case "backspace":
			if query != "" {
				_, n := utf8.DecodeLastRuneInString(query)
				query = query[:len(query)-n]
				cursor = 0
			}
		case "text":
			query += text
			cursor = 0
		}
	}
}

func filterHosts(hosts []string, query string) []string {
	if query == "" {
		return hosts
	}
	words := strings.Fields(strings.ToLower(query))
	result := make([]string, 0)
	for _, host := range hosts {
		lower, matched := strings.ToLower(host), true
		for _, word := range words {
			if !strings.Contains(lower, word) {
				matched = false
				break
			}
		}
		if matched {
			result = append(result, host)
		}
	}
	return result
}

func render(hosts []string, query string, cursor int) {
	width, height := terminalSize(os.Stdout.Fd())
	visible := max(1, height-7)
	start := 0
	if cursor >= visible {
		start = cursor - visible + 1
	}
	end := min(len(hosts), start+visible)
	lineWidth := max(10, width-4)

	var b strings.Builder
	b.WriteString("\x1b[H\x1b[2J\x1b[1;36m  shellector\x1b[0m\n\n")
	fmt.Fprintf(&b, "  \x1b[36m›\x1b[0m Search: %s\x1b[7m \x1b[0m\n\n", query)
	if len(hosts) == 0 {
		b.WriteString("  \x1b[2mNo matching hosts\x1b[0m\n")
	}
	for i := start; i < end; i++ {
		host := truncate(hosts[i], lineWidth)
		if i == cursor {
			fmt.Fprintf(&b, "  \x1b[1;36m❯ %-*s\x1b[0m\n", lineWidth, host)
		} else {
			fmt.Fprintf(&b, "    %-*s\n", lineWidth, host)
		}
	}
	b.WriteString("\n  \x1b[2m↑/↓ navigate  enter connect  esc quit\x1b[0m")
	fmt.Print(b.String())
}

func truncate(s string, width int) string {
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width <= 1 {
		return string(r[:width])
	}
	return string(r[:width-1]) + "…"
}

func readKey() (key, text string, err error) {
	b := make([]byte, 1)
	if _, err = os.Stdin.Read(b); err != nil {
		return
	}
	switch b[0] {
	case 3:
		return "ctrl-c", "", nil
	case 13, 10:
		return "enter", "", nil
	case 127, 8:
		return "backspace", "", nil
	case 27:
		if !inputReady(os.Stdin.Fd(), 40) {
			return "escape", "", nil
		}
		seq := make([]byte, 2)
		if _, err = os.Stdin.Read(seq[:1]); err != nil {
			return "escape", "", nil
		}
		if seq[0] != '[' {
			return "escape", "", nil
		}
		if _, err = os.Stdin.Read(seq[1:]); err != nil {
			return
		}
		if seq[1] == 'A' {
			return "up", "", nil
		}
		if seq[1] == 'B' {
			return "down", "", nil
		}
		return "unknown", "", nil
	default:
		if b[0] >= 32 {
			buf := b
			need := utf8.UTFMax
			for !utf8.FullRune(buf) && len(buf) < need {
				if _, err = os.Stdin.Read(b); err != nil {
					return
				}
				buf = append(buf, b[0])
			}
			return "text", string(buf), nil
		}
	}
	return "unknown", "", nil
}
