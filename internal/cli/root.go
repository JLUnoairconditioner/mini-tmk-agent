package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd 创建根命令
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mini-tmk-agent",
		Short: "A CLI tool for real-time audio transcription and translation",
		Long: `Mini TMK Agent - Transcription & Machine Translation Kit

A powerful CLI tool that combines audio capture, speech recognition, 
and machine translation into a seamless pipeline.

Supports two modes:
1. Stream mode: Real-time microphone transcription and translation
2. Transcript mode: Process audio files and generate transcripts`,
		Version: "1.0.0",
	}

	// 添加子命令
	cmd.AddCommand(NewStreamCmd())
	cmd.AddCommand(NewTranscriptCmd())

	return cmd
}
