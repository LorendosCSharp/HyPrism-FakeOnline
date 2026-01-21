//go:build windows

package game

import (
	"strings"
	"syscall"
	"unsafe"
)

// Windows constants for process creation
const (
	CREATE_NEW_PROCESS_GROUP = 0x00000200
	DETACHED_PROCESS         = 0x00000008
	CREATE_NO_WINDOW         = 0x08000000
)

// getWindowsSysProcAttr returns Windows-specific process attributes
// This hides the console window and detaches the process
func getWindowsSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: CREATE_NEW_PROCESS_GROUP | DETACHED_PROCESS | CREATE_NO_WINDOW,
		HideWindow:    true,
	}
}

// isWindowsProcessRunning checks if a process is running by name using Windows API
// This avoids spawning a visible console window like tasklist would
func isWindowsProcessRunning(processName string) bool {
	// Use CreateToolhelp32Snapshot API to enumerate processes without showing a window
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	createSnapshot := kernel32.NewProc("CreateToolhelp32Snapshot")
	process32First := kernel32.NewProc("Process32FirstW")
	process32Next := kernel32.NewProc("Process32NextW")
	closeHandle := kernel32.NewProc("CloseHandle")

	const TH32CS_SNAPPROCESS = 0x00000002

	type PROCESSENTRY32 struct {
		Size              uint32
		CntUsage          uint32
		ProcessID         uint32
		DefaultHeapID     uintptr
		ModuleID          uint32
		CntThreads        uint32
		ParentProcessID   uint32
		PriorityClassBase int32
		Flags             uint32
		ExeFile           [260]uint16
	}

	snapshot, _, _ := createSnapshot.Call(TH32CS_SNAPPROCESS, 0)
	if snapshot == 0 {
		return false
	}
	defer closeHandle.Call(snapshot)

	var pe PROCESSENTRY32
	pe.Size = uint32(unsafe.Sizeof(pe))

	ret, _, _ := process32First.Call(snapshot, uintptr(unsafe.Pointer(&pe)))
	if ret == 0 {
		return false
	}

	processNameLower := strings.ToLower(processName)

	for {
		exeName := syscall.UTF16ToString(pe.ExeFile[:])
		if strings.ToLower(exeName) == processNameLower {
			return true
		}

		ret, _, _ = process32Next.Call(snapshot, uintptr(unsafe.Pointer(&pe)))
		if ret == 0 {
			break
		}
	}

	return false
}
