//go:build !linux

package app

type terminalState struct{}

func isTerminal(fd uintptr) bool                       { return false }
func makeRaw(fd uintptr) (*terminalState, error)       { return nil, nil }
func restoreTerminal(fd uintptr, state *terminalState) {}
func terminalSize(fd uintptr) (int, int)               { return 80, 24 }
func inputReady(fd uintptr, timeoutMillis int) bool    { return false }
