//go:build !windows

package game

import "syscall"

// getWindowsSysProcAttr returns nil on non-Windows platforms
func getWindowsSysProcAttr() *syscall.SysProcAttr {
	return nil
}

// isWindowsProcessRunning is a stub for non-Windows platforms
// It should never be called on Unix systems
func isWindowsProcessRunning(processName string) bool {
	return false
}
