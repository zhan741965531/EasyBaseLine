package core

import (
	"EasyBaseLine/config"
	"EasyBaseLine/executor"
	"EasyBaseLine/report"
	"EasyBaseLine/utils"
	"EasyBaseLine/web"
	"fmt"
	"github.com/fatih/color"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"runtime"
	"sync"
)

// 定义全局标志变量
var (
	printInfo       = color.New(color.FgGreen).PrintfFunc()
	printError      = color.New(color.FgRed).PrintfFunc()
	printSuccess    = color.New(color.FgHiGreen).PrintfFunc()
	printFail       = color.New(color.FgHiRed).PrintfFunc()
	printHumanCheck = color.New(color.FgHiBlue).PrintfFunc()
)

type Application struct {
	ServerMode bool
	Remote     bool
	File       string
	CheckFile  string
	Ip         string
	Host       string
	Username   string
	Password   string
	Proto      string
	Port       int
}

func NewApplication() *Application {
	app := &Application{}
	return app
}

func (app *Application) Run() {
	if app.ServerMode {
		app.RunServer()
	} else if app.Remote {
		app.executeTests()
	} else {
		app.CheckLocal()
	}
}

func (app *Application) SafePrint(host string, f func(format string, colorFunc ...func(a ...interface{}) string), format string) {
	if host != "" {
		f("[" + host + "] " + format)
	} else {
		f(format)
	}
}

func (app *Application) executeTests() error {

	if app.File != "" {
		return app.executeRemoteTests()
	} else {
		utils.Info("执行远程基线检查\n", color.New(color.FgHiYellow, color.Bold, color.BlinkRapid).SprintFunc())
		app.CheckRemote()
	}

	return nil
}

func (app *Application) executeRemoteTests() error {
	utils.Info("执行远程批量基线检查\n", color.New(color.FgHiYellow, color.Bold, color.BlinkRapid).SprintFunc())

	assets, err := utils.ReadAssetsFromFile(app.File)
	if err != nil {
		return fmt.Errorf("读取资产文件失败：%v", err)
	}

	var wg sync.WaitGroup
	for _, asset := range assets {
		wg.Add(1)
		go func(a config.Asset) {
			defer wg.Done()
			app.CheckAsset(a)
		}(asset)
	}
	wg.Wait()
	return nil
}

func (app *Application) CheckAsset(a config.Asset) {
	testConnect, err := utils.TestConnectToHost(a.Proto, a.Host, a.Username, a.Password, a.Port)
	if err != nil {
		app.SafePrint(a.Host, utils.Error, fmt.Sprintf("远程协议：%s,协议端口：%d,连通性测试失败：%s,原因：%v\n", a.Proto, a.Port, testConnect, err))
		return
	}
	app.SafePrint(a.Host, utils.Info, fmt.Sprintf("远程协议：%s,协议端口：%d,连通性测试成功：%s\n", a.Proto, a.Port, testConnect))
	app.Check(app.CheckFile, a.Host, a.Username, a.Password, a.Port, a.Proto)
}

func (app *Application) Check(configFile string, host, username, password string, port int, proto string) {

	var err error
	var data []byte
	if configFile != "" {
		data, err = ioutil.ReadFile(configFile)

	} else if proto != "" {
		if utils.GetTestSystem(proto) == "windows" {

			data, err = config.BaseLineConfigWindows.ReadFile("checkItems/baseline_config_windows.yaml")
		}
		if utils.GetTestSystem(proto) == "linux" {
			data, err = config.BaseLineConfigLinux.ReadFile("checkItems/baseline_config_linux.yaml")
		}
	} else {
		if runtime.GOOS == "windows" {
			data, err = config.BaseLineConfigWindows.ReadFile("checkItems/baseline_config_windows.yaml")
		}
		if runtime.GOOS == "linux" {
			data, err = config.BaseLineConfigLinux.ReadFile("checkItems/baseline_config_linux.yaml")
		}
	}

	if err != nil {
		utils.Fatal(fmt.Sprintf("读取%s文件时出错: %v\n", configFile, err), color.New(color.FgHiRed).SprintFunc())
	}

	var conf config.Config
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		utils.Fatal(fmt.Sprintf("解析YAML时出错: %v\n", err), color.New(color.FgHiRed).SprintFunc())
	}

	var finalResult config.FinalResult
	for _, item := range conf.Items {
		result := app.ExecuteItemCheck(item, host, username, password, port, proto)
		finalResult.CheckResults = append(finalResult.CheckResults, result)
	}

	finalResult.BasicInfo = conf.BasicInfo
	report.GenerateReport(finalResult)
}

func (app *Application) ExecuteItemCheck(item config.BaselineCheckItem, host, username, password string, port int, proto string) config.CheckResult {
	var cmdOutput string
	var err error

	if app.Remote {
		cmdOutput, err = utils.ExecuteRemoteCommand(host, username, password, item.Query, proto, port)
	} else {
		cmdOutput, err = executor.ExecCommand(item.Query)
	}

	if err != nil {
		utils.Error(fmt.Sprintf("运行查询%s时出错: %v\n", item.UID, err), color.New(color.FgHiRed).SprintFunc())
		return config.CheckResult{} // 返回一个空的 CheckResult
	}

	jsonResult, err := utils.ParseCommandOutput(cmdOutput)
	if err != nil {
		utils.Error(fmt.Sprintf("解析第%s个命令输出时出错: %v,输出为：%s\n", item.UID, err, cmdOutput), color.New(color.FgHiRed).SprintFunc())
		jsonResult = &config.Result{
			Status:  1,
			Outputs: fmt.Sprintf(`无法找到outputs字段,人工检查: %v\n`, err),
		}
	}

	result := config.CheckResult{
		UID:         item.UID,
		Description: item.Description,
		RiskLevel:   item.RiskLevel,
		OutPuts:     jsonResult.Outputs,
		Harm:        item.Harm,
		Solution:    item.Solution,
	}

	switch jsonResult.Status {
	case 0:
		result.Status = "通过"
	case 2:
		result.Status = "人工检查"
	default:
		result.Status = "失败"
	}

	if jsonResult.Status == 0 {
		app.SafePrint(host, utils.Pass, fmt.Sprintf("检查第%s项:%s,检查结果:%s\n", item.UID, item.Description, result.Status))
	} else if jsonResult.Status == 2 {
		app.SafePrint(host, utils.HumanCheck, fmt.Sprintf("检查第%s项:%s,检查结果:%s\n", item.UID, item.Description, result.Status))
	} else {
		app.SafePrint(host, utils.Fail, fmt.Sprintf("检查第%s项:%s,检查结果:%s\n", item.UID, item.Description, result.Status))
	}

	return result
}

func (app *Application) CheckRemote() {
	_, err := utils.TestConnectToHost(app.Proto, app.Host, app.Username, app.Password, app.Port)
	if err != nil {
		app.SafePrint(app.Host, utils.Error, fmt.Sprintf("远程测试连通性失败：%v\n", err))
		return
	}
	app.SafePrint(app.Host, utils.Info, "远程测试连通性成功\n")
	app.Check(app.CheckFile, app.Host, app.Username, app.Password, app.Port, app.Proto)
}

func (app *Application) CheckLocal() {
	utils.Info("执行本地基线检查\n", color.New(color.FgHiYellow, color.Bold, color.BlinkRapid).SprintFunc())
	app.Check(app.CheckFile, "", "", "", 0, "")
}

func (app *Application) RunServer() {
	if app.ServerMode {
		if app.Ip != "" {
			web.Server(app.Ip, "./checkItems")
		} else {
			ipString, err := utils.GetIPAddress()
			if err != nil {
				utils.Fatal(fmt.Sprintf("获取本地ip时出错:%v,需要指定server的ip地址\n", err), color.New(color.FgHiRed).SprintFunc())
			}
			web.Server(ipString, "./checkItems")
		}
	}
}
