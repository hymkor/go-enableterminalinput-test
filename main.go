package main

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	enableVirtualTerminalInput = 0x200

	enableProcessedInput = 0x1
	enableLineInput      = 0x2
	enableEchoInput      = 0x4
	enableWindowInput    = 0x8
	enableMouseInput     = 0x10
	enableInsertMode     = 0x20
	enableQuickEditMode  = 0x40
	enableExtendedFlag   = 0x80

	enableProcessedOutput = 1
	enableWrapAtEolOutput = 2

	keyEvent              = 0x1
	mouseEvent            = 0x2
	windowBufferSizeEvent = 0x4

)

var kernel32 = windows.NewLazyDLL("kernel32.dll")

var (
	procReadConsoleInput = kernel32.NewProc("ReadConsoleInputW")
)

type wchar uint16
type short int16
type dword uint32
type word uint16

type inputRecord struct {
	eventType word
	_         [2]byte
	event     [16]byte
}

type keyEventRecord struct {
	keyDown         int32
	repeatCount     word
	virtualKeyCode  word
	virtualScanCode word
	unicodeChar     wchar
	controlKeyState dword
}

func readConsoleInput(fd uintptr, record *inputRecord) (err error) {
	var w uint32
	r1, _, err := procReadConsoleInput.Call(fd, uintptr(unsafe.Pointer(record)), 1, uintptr(unsafe.Pointer(&w)))
	if r1 == 0 {
		return err
	}
	return nil
}

func mains() error {
	var mode uint32
	err := windows.GetConsoleMode(windows.Stdin, &mode)
	if err != nil {
		return fmt.Errorf("GetConsoleMode: %w", err)
	}
	defer func() {
		windows.SetConsoleMode(windows.Stdin, mode)
	}()

	err = windows.SetConsoleMode(windows.Stdin,
		( mode&^
		enableEchoInput&^
		enableExtendedFlag&^
		enableInsertMode&^
		enableLineInput&^
		enableMouseInput&^
		enableProcessedInput) | enableVirtualTerminalInput )

	if err != nil {
		return err
	}

	for {
		var ir inputRecord
		err := readConsoleInput(uintptr(windows.Stdin), &ir)
		if err != nil {
			return err
		}

		switch ir.eventType {
		case keyEvent:
			kr := (*keyEventRecord)(unsafe.Pointer(&ir.event))
			fmt.Printf("%+v\n", kr)
			if kr.unicodeChar == ' ' {
				return nil
			}
			break
		}
	}
	return nil
}

func main() {
	if err := mains(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
