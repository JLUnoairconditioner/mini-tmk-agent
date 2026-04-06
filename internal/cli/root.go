package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd 创建根命令
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mini-tmk-agent",
		Short: "实时音频转录与翻译 CLI 工具",
		Long: `Mini TMK Agent - 音频转录与机器翻译工具

一个强大的 CLI 工具，将音频采集、语音识别与机器翻译融合为无缝流程。

支持两种模式：
1. Stream 模式：实时麦克风转录与翻译
2. Transcript 模式：处理音频文件并生成转录文本`,
		Version: "1.0.0",
	}

	// 添加子命令
	cmd.AddCommand(NewStreamCmd())
	cmd.AddCommand(NewTranscriptCmd())

	return cmd
}
