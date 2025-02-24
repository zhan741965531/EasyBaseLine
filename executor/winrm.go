package executor

import (
	"fmt"
	"github.com/masterzen/winrm"
	"sync"
)

var clientMutex sync.Mutex

func ExecuteWinRMCommand(remoteHost, username, password, command string, port int) (string, error) {
	var defaultPortWin = 5985
	if port == 0 {
		port = defaultPortWin
	}
	endpoint := winrm.NewEndpoint(remoteHost, port, false, false, nil, nil, nil, 0)
	client, err := winrm.NewClient(endpoint, username, password)
	if err != nil {
		return "", err
	}

	clientMutex.Lock()
	defer clientMutex.Unlock()

	stdout, stderr, exitCode, err := client.RunPSWithString(command, "")
	if err != nil {
		return "", err
	}

	if exitCode != 0 {
		fmt.Errorf("命令退出码：%d", exitCode)
	}
	if stderr != "" {
		fmt.Errorf("命令标准错误：%s", stderr)
	}

	return stdout, nil
}
