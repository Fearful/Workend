//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package main

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
)

// isTerminal returns true if fd is attached to a tty. Best-effort: we just
// check whether stat'ing /dev/tty works and whether stdin is char-special.
func isTerminal(fd int) bool {
	f := os.NewFile(uintptr(fd), "")
	if f == nil {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// readSilent reads a line from /dev/tty with echo disabled via `stty -echo`.
// Falls back to plain line-read if stty isn't available, so the CLI still
// works even on a stripped-down image.
func readSilent(_ int) ([]byte, error) {
	stty, _ := exec.LookPath("stty")
	if stty != "" {
		off := exec.Command(stty, "-echo")
		off.Stdin = os.Stdin
		_ = off.Run()
		defer func() {
			on := exec.Command(stty, "echo")
			on.Stdin = os.Stdin
			_ = on.Run()
		}()
	}
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	return []byte(strings.TrimRight(line, "\r\n")), err
}
