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
		Short: "Transcribe audio files to text",
		Long: `Transcript mode processes audio files and generates transcription and translation.

Supported formats: MP3, WAV, PCM, M4A, FLAC

Example:
  mini-tmk-agent transcript --file audio.mp3 --output result.txt --source-lang zh --target-lang en`,
		RunE: runTranscriptCmd,
	}

	cmd.Flags().String("file", "", "Path to the audio file (required)")
	cmd.Flags().String("output", "", "Path to the output file (required)")
	cmd.Flags().String("source-lang", "zh", "Source language code (zh, en, es, ja)")
	cmd.Flags().String("target-lang", "en", "Target language code (zh, en, es, ja)")
	cmd.Flags().Bool("translate", true, "Enable translation (default: true)")
	cmd.Flags().Bool("verbose", false, "Enable verbose output")

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
		logger.PrintError(fmt.Sprintf("Audio file not found: %s", filePath))
		return fmt.Errorf("file not found: %s", filePath)
	}

	// 验证语言代码
	if !isValidLanguage(sourceLang) || !isValidLanguage(targetLang) {
		logger.PrintError("Invalid language code. Supported: zh, en, es, ja")
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
		logger.PrintError("ASR_API_KEY is not set. Please set the environment variable.")
		return fmt.Errorf("missing ASR_API_KEY")
	}

	if shouldTranslate && cfg.TranslationAPIKey == "" {
		logger.PrintError("TRANSLATION_API_KEY is not set. Please set the environment variable.")
		return fmt.Errorf("missing TRANSLATION_API_KEY")
	}

	logger.PrintInfo(fmt.Sprintf("Processing audio file: %s", filePath))
	logger.PrintInfo(fmt.Sprintf("Source language: %s, Target language: %s",
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
		logger.PrintWarning("File size exceeds 25MB. Will be split into chunks.")
	}

	// 创建 AI 客户端
	asrClient := ai.NewASRClient(cfg.ASRProvider, cfg.ASRAPIKey, cfg.ASRURL, "FunAudioLLM/SenseVoiceSmall")
	var translationClient *ai.TranslationClient
	if shouldTranslate {
		translationClient = ai.NewTranslationClient(cfg.TranslationProvider, cfg.TranslationAPIKey, cfg.TranslationURL, cfg.TranslationModel)
	}

	logger.PrintInfo("Starting transcription...")

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

	logger.PrintSuccess(fmt.Sprintf("Transcription complete (took %.2fs)", asrDuration.Seconds()))
	logger.PrintSourceText(sourceLang, transcribedText)

	var translatedText string
	if shouldTranslate {
		logger.PrintInfo("Starting translation...")
		startTime = time.Now()
		translatedText, err = translationClient.Translate(transcribedText, sourceLang, targetLang)
		if err != nil {
			logger.PrintError(fmt.Sprintf("Failed to translate: %v", err))
			return err
		}
		translationDuration := time.Since(startTime)

		logger.PrintSuccess(fmt.Sprintf("Translation complete (took %.2fs)", translationDuration.Seconds()))
		logger.PrintTargetText(targetLang, translatedText)
	}

	// 生成输出文本
	output := generateOutput(sourceLang, targetLang, transcribedText, translatedText, shouldTranslate, verbose)

	// 保存到文件
	if err := audio.SaveToFile(outputPath, output); err != nil {
		logger.PrintError(fmt.Sprintf("Failed to save output file: %v", err))
		return err
	}

	logger.PrintSuccess(fmt.Sprintf("Results saved to: %s", outputPath))

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
		output += fmt.Sprintf("=== Transcription Report ===\n")
		output += fmt.Sprintf("Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		output += fmt.Sprintf("Source Language: %s\n", getLanguageName(sourceLang))
		output += fmt.Sprintf("Target Language: %s\n", getLanguageName(targetLang))
		output += fmt.Sprintf("\n")
	}

	// 添加转录文本
	output += fmt.Sprintf("=== %s Transcription ===\n", getLanguageName(sourceLang))
	output += transcribedText + "\n\n"

	// 添加翻译文本
	if shouldTranslate {
		output += fmt.Sprintf("=== %s Translation ===\n", getLanguageName(targetLang))
		output += translatedText + "\n"
	}

	return output
}
