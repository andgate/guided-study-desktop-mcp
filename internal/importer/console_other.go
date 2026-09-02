//go:build !windows

package importer

import "os/exec"

// hideConsole does nothing outside Windows.
func hideConsole(_ *exec.Cmd) {}
