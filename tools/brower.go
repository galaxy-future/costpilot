package tools

import (
	"os/exec"
	"runtime"
	"strings"
)

func ShowHtml(fPath string) error {
	sysType := runtime.GOOS
	if strings.EqualFold(sysType, "linux") {
		// LINUX
		return exec.Command(`x-www-browser`, fPath).Start()
	}
	if strings.EqualFold(sysType, "windows") {
		// windows
		return exec.Command(`cmd`, `/c`, `start`, fPath).Start()
	}
	if strings.EqualFold(sysType, "darwin") {
		// mac
		return exec.Command(`open`, fPath).Start()
	}

	return nil
}
