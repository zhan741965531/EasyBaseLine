// Package utils color.go
package utils

import (
	"fmt"
	"github.com/fatih/color"
	"log"
	"os"
	"time"
)

var (
	infoColor       = color.New(color.FgGreen).SprintFunc()
	errorColor      = color.New(color.FgRed).SprintFunc()
	warmColor       = color.New(color.FgYellow).SprintFunc()
	successColor    = color.New(color.FgHiGreen).SprintFunc()
	failColor       = color.New(color.FgHiRed).SprintFunc()
	humanCheckColor = color.New(color.FgHiYellow).SprintFunc()
	timeColor       = color.New(color.FgHiCyan).SprintFunc()
	logFile         *os.File
	fileLogger      *log.Logger
)

func init() {
	var err error
	logFile, err = os.OpenFile("EasyBaseLine.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("无法打开日志文件: %v", err)
	}
	fileLogger = log.New(logFile, "", 0)
}

func printLog(level string, message string, colorFunc func(a ...interface{}) string) {
	// 获取当前时间
	currentTime := timeColor(time.Now().Format("15:04:05"))
	coloredLogMessage := ""
	// 构建带有颜色的日志消息
	switch level {
	case "INFO":
		coloredLogMessage = fmt.Sprintf("[%s] [%s] %s", currentTime, infoColor(level), colorFunc(message))
	case "ERROR":
		coloredLogMessage = fmt.Sprintf("[%s] [%s] %s", currentTime, errorColor(level), colorFunc(message))
	case "WARM":
		coloredLogMessage = fmt.Sprintf("[%s] [%s] %s", currentTime, warmColor(level), colorFunc(message))
	case "FATAL":
		coloredLogMessage = fmt.Sprintf("[%s] [%s] %s", currentTime, errorColor(level), colorFunc(message))
	default:
		coloredLogMessage = fmt.Sprintf("[%s] [%s] %s", currentTime, level, colorFunc(message))
	}

	// 在终端上输出
	fmt.Print(coloredLogMessage)

	// 在日志文件中记录
	fileLogger.Printf("[%s] %s", RemoveColorCodes(currentTime), RemoveColorCodes(message))
}

func Info(message string, colorFunc ...func(a ...interface{}) string) {
	var colorizeFunc func(a ...interface{}) string

	// 如果提供了颜色函数参数，则使用它；否则，默认无颜色
	if len(colorFunc) > 0 {
		colorizeFunc = colorFunc[0]
	} else {
		colorizeFunc = color.New().SprintFunc()
	}

	printLog("INFO", message, colorizeFunc)
}

func Warn(message string, colorFunc ...func(a ...interface{}) string) {
	var colorizeFunc func(a ...interface{}) string

	// 如果提供了颜色函数参数，则使用它；否则，默认黄色
	if len(colorFunc) > 0 {
		colorizeFunc = colorFunc[0]
	} else {
		colorizeFunc = color.New(color.FgYellow).SprintFunc()
	}

	printLog("WARM", message, colorizeFunc)
}

func Error(message string, colorFunc ...func(a ...interface{}) string) {
	var colorizeFunc func(a ...interface{}) string

	// 如果提供了颜色函数参数，则使用它；否则，默认红色
	if len(colorFunc) > 0 {
		colorizeFunc = colorFunc[0]
	} else {
		colorizeFunc = color.New(color.FgRed).SprintFunc()
	}

	printLog("ERROR", message, colorizeFunc)
}
func Fatal(message string, colorFunc ...func(a ...interface{}) string) {
	var colorizeFunc func(a ...interface{}) string

	// 如果提供了颜色函数参数，则使用它；否则，默认红色
	if len(colorFunc) > 0 {
		colorizeFunc = colorFunc[0]
	} else {
		colorizeFunc = color.New(color.FgRed).SprintFunc()
	}

	printLog("FATAL", message, colorizeFunc)
	os.Exit(1)
}

func Pass(message string, colorFunc ...func(a ...interface{}) string) {
	var colorizeFunc func(a ...interface{}) string

	// 如果提供了颜色函数参数，则使用它；否则，默认红色
	if len(colorFunc) > 0 {
		colorizeFunc = colorFunc[0]
	} else {
		colorizeFunc = color.New(color.FgRed).SprintFunc()
	}

	printLog(successColor("Pass"), successColor(message), colorizeFunc)
}

func Fail(message string, colorFunc ...func(a ...interface{}) string) {
	var colorizeFunc func(a ...interface{}) string

	// 如果提供了颜色函数参数，则使用它；否则，默认红色
	if len(colorFunc) > 0 {
		colorizeFunc = colorFunc[0]
	} else {
		colorizeFunc = color.New(color.FgRed).SprintFunc()
	}

	printLog(failColor("NotPass"), failColor(message), colorizeFunc)
}

func HumanCheck(message string, colorFunc ...func(a ...interface{}) string) {
	var colorizeFunc func(a ...interface{}) string

	// 如果提供了颜色函数参数，则使用它；否则，默认红色
	if len(colorFunc) > 0 {
		colorizeFunc = colorFunc[0]
	} else {
		colorizeFunc = color.New(color.FgRed).SprintFunc()
	}

	printLog(humanCheckColor("HumanCheck"), humanCheckColor(message), colorizeFunc)
}
