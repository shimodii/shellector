package app

import (
	"fmt"
	"os"
	"strings"
)

func promptDeleteHost(name string) (bool, error) {
	state, err := makeRaw(os.Stdin.Fd())
	if err != nil {
		return false, err
	}
	fmt.Print("\x1b[?1049h\x1b[?25l")
	defer func() {
		fmt.Print("\x1b[?25h\x1b[?1049l")
		restoreTerminal(os.Stdin.Fd(), state)
	}()

	for {
		fmt.Print(formatDeletePrompt(name))
		key, text, err := readKey()
		if err != nil {
			return false, err
		}
		if key == "escape" || key == "ctrl-c" || (key == "text" && strings.EqualFold(text, "n")) {
			return false, nil
		}
		if key == "text" && strings.EqualFold(text, "y") {
			return true, nil
		}
	}
}

func formatDeletePrompt(name string) string {
	return fmt.Sprintf("\x1b[H\x1b[2J\x1b[1;31m  Delete SSH host\x1b[0m\n\n  Delete \x1b[1m%s\x1b[0m from your SSH config?\n  This cannot be undone.\n\n  \x1b[2mY confirm  N/esc cancel\x1b[0m", name)
}
