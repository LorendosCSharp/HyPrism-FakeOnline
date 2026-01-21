//go:build windows

package util

import (
	"os/exec"
	"syscall"
)

// HideConsoleWindow hides the console window for commands on Windows
func HideConsoleWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}
