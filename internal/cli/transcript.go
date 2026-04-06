package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mini-tmk-agent/internal/ai"
	"mini-tmk-agent/internal/audio"
	"mini-tmk-agent/internal/config"
	"mini-tmk-agent/pkg/logger"

	"github.com/spf13/cobra"
)

// NewTranscriptCmd 创建 transcript 子命令
func NewTranscriptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transcript",
		Short: "将音频文件转录为文本",
		Long: `Transcript 模式处理音频文件，并生成转录和翻译结果。

支持格式：MP3、WAV、PCM、M4A、FLAC

示例：
  mini-tmk-agent transcript --file audio.mp3 --output result.txt --source-lang zh --target-lang en`,
		RunE: runTranscriptCmd,
	}

	cmd.Flags().String("file", "", "音频文件路径（必填）")
	cmd.Flags().String("output", "", "输出文件路径（必填）")
	cmd.Flags().String("source-lang", "zh", "源语言代码 (zh, en, es, ja)")
	cmd.Flags().String("target-lang", "en", "目标语言代码 (zh, en, es, ja)")
	cmd.Flags().Bool("translate", true, "是否启用翻译（默认: true）")
	cmd.Flags().Bool("verbose", false, "启用详细输出")

	cmd.MarkFlagRequired("file")
	cmd.MarkFlagRequired("output")

	return cmd
}

func runTranscriptCmd(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	outputPath, _ := cmd.Flags().GetString("output")
	sourceLang, _ := cmd.Flags().GetString("source-lang")
	targetLang, _ := cmd.Flags().GetString("target-lang")
	shouldTranslate, _ := cmd.Flags().GetBool("translate")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// 验证文件存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.PrintError(fmt.Sprintf("未找到音频文件：%s", filePath))
		return fmt.Errorf("file not found: %s", filePath)
	}

	// 验证语言代码
	if !isValidLanguage(sourceLang) || !isValidLanguage(targetLang) {
		logger.PrintError("语言代码无效。支持：zh, en, es, ja")
		return fmt.Errorf("invalid language code")
	}

	// 验证输出目录
	outputDir := filepath.Dir(outputPath)
	if outputDir != "." && outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			logger.PrintError(fmt.Sprintf("Failed to create output directory: %v", err))
			return err
		}
	}

	cfg := config.LoadConfig()

	// 验证 API 密钥
	if cfg.ASRAPIKey == "" {
		logger.PrintError("未设置 ASR_API_KEY，请设置环境变量。")
		return fmt.Errorf("missing ASR_API_KEY")
	}

	if shouldTranslate && cfg.TranslationAPIKey == "" {
		logger.PrintError("未设置 TRANSLATION_API_KEY，请设置环境变量。")
		return fmt.Errorf("missing TRANSLATION_API_KEY")
	}

	logger.PrintInfo(fmt.Sprintf("处理音频文件：%s", filePath))
	logger.PrintInfo(fmt.Sprintf("源语言：%s，目标语言：%s",
		getLanguageName(sourceLang), getLanguageName(targetLang)))

	// 读取音频文件
	audioFile, err := audio.NewAudioFile(filePath)
	if err != nil {
		logger.PrintError(fmt.Sprintf("Failed to load audio file: %v", err))
		return err
	}

	fileSize := audioFile.GetSize()
	logger.PrintInfo(fmt.Sprintf("Audio file size: %.2f MB", float64(fileSize)/1024/1024))

	// 检查文件大小（OpenAI Whisper API 限制 25MB）
	if fileSize > 25*1024*1024 {
		logger.PrintWarning("文件大小超过 25MB，将拆分成多个块处理。")
	}

	// 创建 AI 客户端
	asrClient := ai.NewASRClient(cfg.ASRProvider, cfg.ASRAPIKey, cfg.ASRURL, "FunAudioLLM/SenseVoiceSmall")
	var translationClient *ai.TranslationClient
	if shouldTranslate {
		translationClient = ai.NewTranslationClient(cfg.TranslationProvider, cfg.TranslationAPIKey, cfg.TranslationURL, cfg.TranslationModel)
	}

	logger.PrintInfo("开始转录...")

	// 获取音频数据
	audioData := audioFile.GetData()

	// 调用 ASR
	startTime := time.Now()
	transcribedText, err := asrClient.Transcribe(audioData, sourceLang)
	if err != nil {
		logger.PrintError(fmt.Sprintf("Failed to transcribe: %v", err))
		return err
	}
	asrDuration := time.Since(startTime)

	logger.PrintSuccess(fmt.Sprintf("转录完成（耗时 %.2fs）", asrDuration.Seconds()))
	logger.PrintSourceText(sourceLang, transcribedText)

	var translatedText string
	if shouldTranslate {
		logger.PrintInfo("开始翻译...")
		startTime = time.Now()
		translatedText, err = translationClient.Translate(transcribedText, sourceLang, targetLang)
		if err != nil {
			logger.PrintError(fmt.Sprintf("Failed to translate: %v", err))
			return err
		}
		translationDuration := time.Since(startTime)

		logger.PrintSuccess(fmt.Sprintf("翻译完成（耗时 %.2fs）", translationDuration.Seconds()))
		logger.PrintTargetText(targetLang, translatedText)
	}

	// 生成输出文本
	output := generateOutput(sourceLang, targetLang, transcribedText, translatedText, shouldTranslate, verbose)

	// 保存到文件
	if err := audio.SaveToFile(outputPath, output); err != nil {
		logger.PrintError(fmt.Sprintf("Failed to save output file: %v", err))
		return err
	}

	logger.PrintSuccess(fmt.Sprintf("结果已保存到：%s", outputPath))

	// 打印文件统计
	if verbose {
		logger.PrintInfo(fmt.Sprintf("Transcribed text length: %d characters", len(transcribedText)))
		if shouldTranslate {
			logger.PrintInfo(fmt.Sprintf("Translated text length: %d characters", len(translatedText)))
		}
	}

	return nil
}

func generateOutput(sourceLang, targetLang, transcribedText, translatedText string, shouldTranslate bool, verbose bool) string {
	output := ""

	// 添加元数据
	if verbose {
		output += fmt.Sprintf("=== 转录报告 ===\n")
		output += fmt.Sprintf("生成时间：%s\n", time.Now().Format("2006-01-02 15:04:05"))
		output += fmt.Sprintf("源语言：%s\n", getLanguageName(sourceLang))
		output += fmt.Sprintf("目标语言：%s\n", getLanguageName(targetLang))
		output += fmt.Sprintf("\n")
	}

	// 添加转录文本
	output += fmt.Sprintf("=== %s 转录 ===\n", getLanguageName(sourceLang))
	output += transcribedText + "\n\n"

	// 添加翻译文本
	if shouldTranslate {
		output += fmt.Sprintf("=== %s 翻译 ===\n", getLanguageName(targetLang))
		output += translatedText + "\n"
	}

	return output
}
