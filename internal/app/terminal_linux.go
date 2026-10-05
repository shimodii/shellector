//go:build linux

package app

import (
	"syscall"
	"unsafe"
)

type terminalState syscall.Termios

func isTerminal(fd uintptr) bool {
	var term syscall.Termios
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TCGETS, uintptr(unsafe.Pointer(&term)), 0, 0, 0)
	return errno == 0
}

func makeRaw(fd uintptr) (*terminalState, error) {
	var old syscall.Termios
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TCGETS, uintptr(unsafe.Pointer(&old)), 0, 0, 0)
	if errno != 0 {
		return nil, errno
	}
	raw := rawTermios(old)
	_, _, errno = syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TCSETS, uintptr(unsafe.Pointer(&raw)), 0, 0, 0)
	if errno != 0 {
		return nil, errno
	}
	state := terminalState(old)
	return &state, nil
}

func rawTermios(old syscall.Termios) syscall.Termios {
	raw := old
	raw.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	raw.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	raw.Cflag &^= syscall.CSIZE | syscall.PARENB
	raw.Cflag |= syscall.CS8
	raw.Cc[syscall.VMIN], raw.Cc[syscall.VTIME] = 1, 0
	return raw
}

func restoreTerminal(fd uintptr, state *terminalState) {
	term := syscall.Termios(*state)
	syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TCSETS, uintptr(unsafe.Pointer(&term)), 0, 0, 0)
}

func terminalSize(fd uintptr) (int, int) {
	type winsize struct{ Row, Col, Xpixel, Ypixel uint16 }
	var ws winsize
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&ws)), 0, 0, 0)
	if errno != 0 || ws.Row == 0 || ws.Col == 0 {
		return 80, 24
	}
	return int(ws.Col), int(ws.Row)
}

func inputReady(fd uintptr, timeoutMillis int) bool {
	var set syscall.FdSet
	set.Bits[fd/64] |= 1 << (fd % 64)
	timeout := syscall.Timeval{Usec: int64(timeoutMillis * 1000)}
	n, err := syscall.Select(int(fd)+1, &set, nil, nil, &timeout)
	return err == nil && n > 0
}
