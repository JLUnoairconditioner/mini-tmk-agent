package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"mini-tmk-agent/internal/ai"
	"mini-tmk-agent/internal/audio"
	"mini-tmk-agent/internal/config"
	"mini-tmk-agent/pkg/logger"

	"github.com/spf13/cobra"
)

// NewStreamCmd 创建 stream 子命令
func NewStreamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Real-time audio streaming with live transcription and translation",
		Long: `Stream mode enables real-time transcription and translation from your microphone.

It listens to your microphone continuously and performs live translation
of your speech to the target language.`,
		RunE: runStreamCmd,
	}

	cmd.Flags().String("source-lang", "zh", "Source language code (zh, en, es, ja)")
	cmd.Flags().String("target-lang", "en", "Target language code (zh, en, es, ja)")
	cmd.Flags().Bool("verbose", false, "Enable verbose output")

	return cmd
}

func runStreamCmd(cmd *cobra.Command, args []string) error {
	sourceLang, _ := cmd.Flags().GetString("source-lang")
	targetLang, _ := cmd.Flags().GetString("target-lang")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// 验证语言代码
	if !isValidLanguage(sourceLang) || !isValidLanguage(targetLang) {
		logger.PrintError(fmt.Sprintf("Invalid language code. Supported: zh, en, es, ja"))
		return fmt.Errorf("invalid language code")
	}

	if sourceLang == targetLang {
		logger.PrintWarning("Source and target languages are the same")
	}

	cfg := config.LoadConfig()

	// 验证 API 密钥
	if cfg.ASRAPIKey == "" {
		logger.PrintError("ASR_API_KEY is not set. Please set the environment variable.")
		return fmt.Errorf("missing ASR_API_KEY")
	}

	if cfg.TranslationAPIKey == "" {
		logger.PrintError("TRANSLATION_API_KEY is not set. Please set the environment variable.")
		return fmt.Errorf("missing TRANSLATION_API_KEY")
	}

	logger.PrintInfo(fmt.Sprintf("Starting stream mode: %s -> %s", sourceLang, targetLang))
	logger.PrintInfo("Initializing microphone...")

	// 创建录音机
	recorder, err := audio.NewRecorder(cfg.SampleRate, cfg.NumChannels, cfg.FramesPerBuffer)
	if err != nil {
		logger.PrintError(fmt.Sprintf("Failed to initialize recorder: %v", err))
		return err
	}
	defer recorder.Close()

	// 启动录音
	if err := recorder.Start(); err != nil {
		logger.PrintError(fmt.Sprintf("Failed to start recording: %v", err))
		return err
	}

	logger.PrintSuccess("Microphone initialized. Listening...")
	logger.PrintInfo("Press Ctrl+C to stop")

	// 创建 AI 客户端
	asrClient := ai.NewASRClient(cfg.ASRProvider, cfg.ASRAPIKey, cfg.ASRURL, cfg.ASRModel)
	translationClient := ai.NewTranslationClient(cfg.TranslationProvider, cfg.TranslationAPIKey, cfg.TranslationURL, cfg.TranslationModel)

	// 创建 VAD
	vad := audio.NewVAD(cfg.VADThreshold, cfg.VADSilenceDuration, cfg.SampleRate)

	// 创建上下文和信号处理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 创建通道
	audioChan := make(chan []float32, 10)
	textChan := make(chan string, 10)

	var wg sync.WaitGroup

	// Goroutine 1: 音频采集
	wg.Add(1)
	go func() {
		defer wg.Done()
		var audioBuffer [][]float32
		var hasDetectedSpeech bool
		var prevIsActive bool

		for {
			select {
			case <-ctx.Done():
				return
			default:
				frame, err := recorder.Read()
				if err != nil {
					logger.PrintError(fmt.Sprintf("Error reading audio: %v", err))
					return
				}

				audioBuffer = append(audioBuffer, frame)

				// 检测语音活动
				isActive, energy := vad.DetectActivity(frame, len(frame))

				if isActive {
					hasDetectedSpeech = true
					if verbose {
						logger.PrintInfo(fmt.Sprintf("🎤 Detecting speech... energy=%.5f threshold=%.5f", energy, cfg.VADThreshold))
					}
				} else if verbose {
					if !prevIsActive {
						logger.PrintInfo("🔇 当前没有检测到声音，等待静音结束...")
					}
					logger.PrintInfo(fmt.Sprintf("    energy=%.5f threshold=%.5f", energy, cfg.VADThreshold))
				}

				// 只要当前段曾听到过声音，并且 VAD 认为当前已经进入静音状态，就发送整段音频
				if !isActive && hasDetectedSpeech && len(audioBuffer) > 0 {
					// 计算音频时长
					totalSamples := 0
					for _, buf := range audioBuffer {
						totalSamples += len(buf)
					}
					audioDurationMs := (totalSamples * 1000) / 16000 // ms
					
					logger.PrintInfo(fmt.Sprintf("✓ VAD 判断为静音，发送 %.1f 秒的音频进行识别...", float64(audioDurationMs)/1000))
					
					// 合并音频帧
					var fullAudio []float32
					for _, buf := range audioBuffer {
						fullAudio = append(fullAudio, buf...)
					}

					select {
					case <-ctx.Done():
						return
					case audioChan <- fullAudio:
						audioBuffer = [][]float32{}
						hasDetectedSpeech = false
						vad.Reset()
					}
				}
				prevIsActive = isActive
			}
		}
	}()

	// Goroutine 2: ASR 识别
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case audioFrame := <-audioChan:
				if audioFrame == nil {
					return
				}

				// 将浮点数音频转换为字节
				audioBytes := convertFloat32ToBytes(audioFrame)

				if verbose {
					logger.PrintInfo(fmt.Sprintf("Sending %d bytes to ASR...", len(audioBytes)))
				}

				// 调用 ASR
				text, err := asrClient.Transcribe(audioBytes, sourceLang)
				if err != nil {
					logger.PrintError(fmt.Sprintf("ASR error: %v", err))
					continue
				}

				if text == "" {
					continue
				}

				logger.PrintSourceText(sourceLang, text)

				select {
				case <-ctx.Done():
					return
				case textChan <- text:
				}
			}
		}
	}()

	// Goroutine 3: 翻译
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case sourceText := <-textChan:
				if sourceText == "" {
					return
				}

				if verbose {
					logger.PrintInfo("Translating...")
				}

				// 调用翻译 API
				translatedText, err := translationClient.Translate(sourceText, sourceLang, targetLang)
				if err != nil {
					logger.PrintError(fmt.Sprintf("Translation error: %v", err))
					continue
				}

				if translatedText != "" {
					logger.PrintTargetText(targetLang, translatedText)
				}
			}
		}
	}()

	// 等待信号
	<-sigChan
	logger.PrintInfo("\nShutting down...")
	recorder.Stop()
	cancel()
	close(audioChan)
	close(textChan)

	wg.Wait()
	logger.PrintSuccess("Done!")

	return nil
}

func isValidLanguage(lang string) bool {
	validLanguages := map[string]bool{
		"zh": true, // 中文
		"en": true, // 英文
		"es": true, // 西班牙文
		"ja": true, // 日文
	}
	return validLanguages[lang]
}

func getLanguageName(lang string) string {
	names := map[string]string{
		"zh": "Chinese",
		"en": "English",
		"es": "Spanish",
		"ja": "Japanese",
	}
	if name, ok := names[lang]; ok {
		return name
	}
	return lang
}

// convertFloat32ToBytes 将 float32 音频数据转换为 WAV 格式
func convertFloat32ToBytes(audio []float32) []byte {
	// 首先转换为 int16 PCM 数据
	pcmData := make([]byte, len(audio)*2)
	for i, v := range audio {
		// 将 float32 转换为 int16
		sample := int16(v * 32767)
		pcmData[i*2] = byte(sample)
		pcmData[i*2+1] = byte(sample >> 8)
	}

	// 创建 WAV 文件头
	sampleRate := 16000
	channels := 1
	bytesPerSample := 2

	wavHeader := make([]byte, 44)

	// RIFF header
	copy(wavHeader[0:4], []byte("RIFF"))

	// File size - 8
	fileSize := len(pcmData) + 36
	wavHeader[4] = byte(fileSize)
	wavHeader[5] = byte(fileSize >> 8)
	wavHeader[6] = byte(fileSize >> 16)
	wavHeader[7] = byte(fileSize >> 24)

	// WAVE header
	copy(wavHeader[8:12], []byte("WAVE"))

	// fmt subchunk
	copy(wavHeader[12:16], []byte("fmt "))
	wavHeader[16] = 16             // Subchunk1Size
	wavHeader[20] = 1              // AudioFormat (PCM)
	wavHeader[22] = byte(channels) // NumChannels

	// Sample rate
	wavHeader[24] = byte(sampleRate)
	wavHeader[25] = byte(sampleRate >> 8)
	wavHeader[26] = byte(sampleRate >> 16)
	wavHeader[27] = byte(sampleRate >> 24)

	// Byte rate
	byteRate := sampleRate * channels * bytesPerSample
	wavHeader[28] = byte(byteRate)
	wavHeader[29] = byte(byteRate >> 8)
	wavHeader[30] = byte(byteRate >> 16)
	wavHeader[31] = byte(byteRate >> 24)

	// Block align
	wavHeader[32] = byte(channels * bytesPerSample)

	// Bits per sample
	wavHeader[34] = 16

	// data subchunk
	copy(wavHeader[36:40], []byte("data"))
	wavHeader[40] = byte(len(pcmData))
	wavHeader[41] = byte(len(pcmData) >> 8)
	wavHeader[42] = byte(len(pcmData) >> 16)
	wavHeader[43] = byte(len(pcmData) >> 24)

	// 合并 WAV 头和 PCM 数据
	result := append(wavHeader, pcmData...)
	return result
}
