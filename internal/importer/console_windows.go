package importer

import (
	"os/exec"
	"syscall"
)

// hideConsole keeps the converter window off screen.
func hideConsole(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
