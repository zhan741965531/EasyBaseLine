package executor

import (
	"errors"
	"os/exec"
	"runtime"
)

func ExecCommand(command string) (string, error) {
	var Cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		Cmd = exec.Command("powershell", "-NoProfile", "[Console]::OutputEncoding = [System.Text.Encoding]::UTF8\r\n"+command)
	case "linux":
		Cmd = exec.Command("/bin/bash", "-c", command)
	default:
		return "", errors.New("unsupported operating system")
	}

	result, err := Cmd.CombinedOutput()
	return string(result), err
}
