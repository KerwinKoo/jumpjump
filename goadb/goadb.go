package goadb

import (
	"fmt"
	"os/exec"
)

// PixPoint device Physical size define
type PixPoint struct {
	X int
	Y int
}

// common defines
var (
	CommonPySize = &PixPoint{720, 1280}
)

// ExecAdb exec adb commands
func ExecAdb(cmdStr string) (string, error) {
	cmd := exec.Command("sh", "-c", cmdStr)

	out, err := cmd.Output()

	if err != nil {
		return "", err
	}

	return string(out), nil
}

// ScreenShot do screenshot by adb
func ScreenShot(saveingPath string) error {
	_, err := ExecAdb("adb shell /system/bin/screencap -p /sdcard/screenshot.png")
	if err != nil {
		return err
	}

	pullCmd := fmt.Sprintf("adb pull /sdcard/screenshot.png %s", saveingPath)
	_, err = ExecAdb(pullCmd)
	if err != nil {
		return err
	}

	return nil
}

// LongPress do long press by adb
func LongPress(psFrom, psEnd *PixPoint, dur int) error {
	cmdStr := fmt.Sprintf("adb shell input swipe %d %d %d %d %d",
		psFrom.X, psFrom.Y, psEnd.X, psEnd.Y, dur)

	_, err := ExecAdb(cmdStr)
	if err != nil {
		return err
	}

	return nil
}
