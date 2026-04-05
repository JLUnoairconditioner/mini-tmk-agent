package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 应用配置结构体
type Config struct {
	// ASR 配置
	ASRProvider string // "openai" 或 "siliconflow"
	ASRAPIKey   string
	ASRURL      string
	ASRModel    string
	// Translation 配置
	TranslationProvider string // "openai", "deepseek", "qwen" 等
	TranslationAPIKey   string
	TranslationURL      string
	TranslationModel    string
	// Audio 配置
	SampleRate      int // 音频采样率，默认 16000
	BitDepth        int // 位深度，默认 16
	NumChannels     int // 通道数，默认 1
	FramesPerBuffer int // 每次读取的帧数，默认 2048

	// VAD 配置
	VADEnabled         bool    // 是否启用 VAD
	VADThreshold       float32 // 能量阈值
	VADSilenceDuration int     // 静音持续时间（毫秒）
}

// LoadConfig 从环境变量加载配置
// func LoadConfig() *Config {
// 	// 尝试加载 .env 文件（如果存在）
// 	_ = godotenv.Load()

// 	cfg := &Config{
// 		ASRProvider:         getEnv("ASR_PROVIDER", "openai"),
// 		ASRAPIKey:           getEnv("ASR_API_KEY", ""),
// 		ASRURL:              getEnv("ASR_URL", "https://api.openai.com/v1/audio/transcriptions"),
// 		TranslationProvider: getEnv("TRANSLATION_PROVIDER", "openai"),
// 		TranslationAPIKey:   getEnv("TRANSLATION_API_KEY", ""),
// 		TranslationURL:      getEnv("TRANSLATION_URL", "https://api.openai.com/v1/chat/completions"),
// 		SampleRate:          16000,
// 		BitDepth:            16,
// 		NumChannels:         1,
// 		FramesPerBuffer:     2048,
// 		VADEnabled:          getEnvBool("VAD_ENABLED", true),
// 		VADThreshold:        0.02,
// 		VADSilenceDuration:  500,
// 	}

// 	return cfg
// }
// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
    // 尝试加载 .env 文件（如果存在）
    _ = godotenv.Load()

    cfg := &Config{
        // 将默认 Provider 改为 siliconflow
        ASRProvider:         getEnv("ASR_PROVIDER", "siliconflow"),
        ASRAPIKey:           getEnv("ASR_API_KEY", ""),
        // 替换为硅基流动的语音识别地址
        ASRURL:              getEnv("ASR_URL", "https://api.siliconflow.cn/v1/audio/transcriptions"),
        // 新增：默认使用硅基流动支持的 SenseVoice 语音模型
        ASRModel:            getEnv("ASR_MODEL", "FunAudioLLM/SenseVoiceSmall"), 
        
        TranslationProvider: getEnv("TRANSLATION_PROVIDER", "deepseek"),
        TranslationAPIKey:   getEnv("TRANSLATION_API_KEY", ""),
        // DeepSeek 翻译服务地址
        TranslationURL:      getEnv("TRANSLATION_URL", "https://api.deepseek.com/chat/completions"),
        // 翻译模型
        TranslationModel:    getEnv("TRANSLATION_MODEL", "deepseek-chat"),

        SampleRate:          16000,
        BitDepth:            16,
        NumChannels:         1,
        FramesPerBuffer:     8192, // 增加到 8192 (512ms/次) 拉长切片发送间隔
        VADEnabled:          getEnvBool("VAD_ENABLED", true),
        VADThreshold:        getEnvFloat("VAD_THRESHOLD", 0.02), // 默认阈值降低到 0.02，减少说话时被误判为静音
        
        // 缩短静音检测时间到 1.5 秒，减少说话结束后等待时间
        VADSilenceDuration:  getEnvInt("VAD_SILENCE_DURATION", 1500),
    }

    return cfg
}
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	return value == "true" || value == "1" || value == "yes"
}

func getEnvFloat(key string, fallback float32) float32 {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return fallback
	}
	return float32(f)
}

func getEnvInt(key string, fallback int) int {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return i
}
