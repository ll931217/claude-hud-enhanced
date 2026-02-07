package terminal

import (
	"os"
	"syscall"
	"unsafe"
)

// Size represents terminal dimensions
type Size struct {
	Columns int
	Rows    int
}

// GetSize retrieves the terminal size using TIOCGWINSZ
func GetSize() Size {
	ws := struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}{}

	_, _, _ = syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(os.Stdout.Fd()),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ws)),
	)

	return Size{
		Columns: int(ws.Col),
		Rows:    int(ws.Row),
	}
}

// AvailableWidth returns available columns (with safety margin)
func AvailableWidth() int {
	size := GetSize()
	// Leave 2 columns margin on each side
	if size.Columns <= 4 {
		return 0
	}
	return size.Columns - 4
}

// AvailableRows returns available rows (with safety margin)
func AvailableRows() int {
	size := GetSize()
	// Leave 1 row margin
	if size.Rows <= 1 {
		return 0
	}
	return size.Rows - 1
}
