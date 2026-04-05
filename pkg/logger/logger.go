package logger

import (
	"fmt"

	"github.com/pterm/pterm"
)

// PrintSourceText 打印源语言文本（绿色）
func PrintSourceText(lang, text string) {
	pterm.Info.Println(pterm.Green(fmt.Sprintf("[%s] %s", lang, text)))
}

// PrintTargetText 打印目标语言文本（蓝色）
func PrintTargetText(lang, text string) {
	pterm.Info.Println(pterm.Blue(fmt.Sprintf("[%s] %s", lang, text)))
}

// PrintSuccess 打印成功消息
func PrintSuccess(msg string) {
	pterm.Success.Println(msg)
}

// PrintError 打印错误消息
func PrintError(msg string) {
	pterm.Error.Println(msg)
}

// PrintWarning 打印警告消息
func PrintWarning(msg string) {
	pterm.Warning.Println(msg)
}

// PrintInfo 打印信息消息
func PrintInfo(msg string) {
	pterm.Info.Println(msg)
}
