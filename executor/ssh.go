package executor

import (
	"fmt"
	"golang.org/x/crypto/ssh"
)

func ExecuteSSHCmd(host string, port int, username, password, cmd string) (string, error) {

	var defaultPortSSH = 22

	if port == 0 {
		port = defaultPortSSH
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
	if err != nil {
		return "", fmt.Errorf("failed to dial:%v", err)
	}

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session:%v", err)
	}

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to run command:%v", err)
	}

	return string(output), nil
}
