package utils

import (
	"EasyBaseLine/config"
	"EasyBaseLine/executor"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/masterzen/winrm"
	"golang.org/x/crypto/ssh"
	"math/rand"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func GetIPAddress() (string, error) {
	hostName, err := os.Hostname()
	if err != nil {
		return "", err
	}

	adds, err := net.LookupHost(hostName)
	if err != nil {
		return "", err
	}

	for _, addr := range adds {
		ip := net.ParseIP(addr).To4()
		if ip != nil {
			return ip.String(), nil
		}
	}

	return "", errors.New("无法确定IP地址")
}

func GetTestSystem(proto string) string {
	if proto == "winrm" {
		return "windows"
	} else if proto == "ssh" {
		return "linux"
	}
	return ""
}

func ParseCommandOutput(cmdOutput string) (*config.Result, error) {
	pattern := `(?s){\s*"outputs"\s*:\s*".*?"\s*,\s*"status"\s*:\s*[012]\s*}`
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("正则表达式编译失败: %v", err)
	}

	match := regex.FindString(cmdOutput)
	if match == "" {
		return nil, fmt.Errorf("无法匹配命令输出")
	}

	var jsonResult config.Result
	err = json.Unmarshal([]byte(match), &jsonResult)
	//if err != nil {
	//	return nil, fmt.Errorf("解析JSON失败: %v", err)
	//}
	if err != nil {
		jsonResult.Status = 1
		var outputs string
		var statusIndex int
		re1 := regexp.MustCompile(`\s+`)
		cmdOutput := re1.ReplaceAllString(cmdOutput, "")
		cmdOutput = strings.ReplaceAll(cmdOutput, "，", ",")
		startIndex := strings.Index(cmdOutput, `"outputs":"`)
		endIndex := strings.Index(cmdOutput, `","status":`)
		statusIndex = endIndex + len(`","status":`)
		if startIndex == -1 || endIndex == -1 {
			cmdOutput = `{"outputs":"无法匹配到outputs输出,需要人工检查","status":2}`
			startIndex = strings.Index(cmdOutput, `"outputs": "`)
			endIndex = strings.Index(cmdOutput, `","status":`)
			statusIndex = endIndex + len(`","status":`)
		}
		outputs = cmdOutput[startIndex+len(`"outputs":"`) : endIndex]
		statusStr := cmdOutput[statusIndex : len(cmdOutput)-2]
		re2 := regexp.MustCompile("[012]")
		matches := re2.FindString(strings.TrimSpace(statusStr))
		num, _ := strconv.Atoi(matches)
		jsonResult.Status = num
		jsonResult.Outputs = outputs
	}
	return &jsonResult, nil
}

func TestConnectToHost(proto, remoteHost, username, password string, port int) (string, error) {
	var defaultPortWin = 5985
	var defaultPortSSH = 22

	if proto == "winrm" && port == 0 {
		port = defaultPortWin
	} else if proto == "ssh" && port == 0 {
		port = defaultPortSSH
	}

	switch proto {
	case "winrm":
		endpoint := winrm.NewEndpoint(remoteHost, port, false, false, nil, nil, nil, 0)
		client, err := winrm.NewClient(endpoint, username, password)
		if err != nil {
			return "NotPass", err
		}

		// 执行一个简单的命令来测试连接
		testCmd := "echo 'test'"
		_, _, exitCode, err := client.RunPSWithString(testCmd, "")
		if err != nil || exitCode != 0 {
			return "NotPass", fmt.Errorf("failed to execute test command via WinRM: %v", err)
		}

		return "Pass", nil

	case "ssh":
		config := &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{
				ssh.Password(password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		_, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", remoteHost, port), config)
		if err != nil {
			return "NotPass", fmt.Errorf("failed to dial:%v", err)
		}
		// 考虑执行一个简单的命令来测试连接
		return "Pass", nil

	default:
		return "", errors.New("不支持的协议，只支持ssh和winrm")
	}
}

func ReadAssetsFromFile(file string) ([]config.Asset, error) {
	var assets []config.Asset

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 4 {
			asset := config.Asset{
				Host:     fields[0],
				Username: fields[1],
				Password: fields[2],
				Proto:    fields[3],
			}
			assets = append(assets, asset)
		} else {
			port, _ := strconv.Atoi(fields[4])
			asset := config.Asset{
				Host:     fields[0],
				Username: fields[1],
				Password: fields[2],
				Proto:    fields[3],
				Port:     port,
			}
			assets = append(assets, asset)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return assets, nil
}

func ExecuteRemoteCommand(remoteHost, username, password, command, proto string, port int) (string, error) {
	if proto == "winrm" {
		return executor.ExecuteWinRMCommand(remoteHost, username, password, command, port)
	} else if proto == "ssh" {
		return executor.ExecuteSSHCmd(remoteHost, port, username, password, command)
	}

	return "", fmt.Errorf("不支持的协议: %s", proto)
}

const (
	usernameCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	passwordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:'\",.<>?/"
	usernameLength  = 6
	passwordLength  = 12
)

func randomString(charset string, length int) string {
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func GenerateAccount() (string, string) {
	username := randomString(usernameCharset, usernameLength)
	password := randomString(passwordCharset, passwordLength)
	return username, password
}

func CheckAndCreateFolder(folderName string) {
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		Info(fmt.Sprintf("文件夹 '%s' 不存在，正在创建...\n", folderName))
		err := os.MkdirAll(folderName, os.ModePerm)
		if err != nil {
			Fatal(fmt.Sprintf("无法创建文件夹 '%s': %v\n", folderName, err))
			return
		}
	}
}

func RemoveColorCodes(input string) string {
	// 正则表达式匹配颜色符号
	re := regexp.MustCompile("\x1b\\[[0-9;]+m")

	// 替换颜色符号为空字符串
	cleanOutput := re.ReplaceAllString(input, "")

	return cleanOutput
}
