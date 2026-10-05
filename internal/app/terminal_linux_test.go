//go:build linux

package app

import (
	"syscall"
	"testing"
)

func TestRawTerminalPreservesOutputProcessing(t *testing.T) {
	original := syscall.Termios{Oflag: syscall.OPOST | syscall.ONLCR}
	raw := rawTermios(original)
	if raw.Oflag&syscall.OPOST == 0 || raw.Oflag&syscall.ONLCR == 0 {
		t.Fatal("raw mode must preserve newline output processing")
	}
}
