package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	//"mini-tmk-agent/internal/config"
)

// ASRClient ASR (自动语音识别) 客户端
type ASRClient struct {
	provider string
	apiKey   string
	url      string
	client   *http.Client
	model    string
}

// NewASRClient 创建 ASR 客户端
func NewASRClient(provider, apiKey, url, model string) *ASRClient {
	return &ASRClient{
		provider: provider,
		apiKey:   apiKey,
		url:      url,
		client:   &http.Client{},
		model:    model,
	}
}

// TranscribeOpenAI 使用 OpenAI Whisper API 转录音频
func (a *ASRClient) TranscribeOpenAI(audioData []byte, language string) (string, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// 添加语言参数
	_ = writer.WriteField("language", language)
	_ = writer.WriteField("model", a.model)

	// 添加音频文件
	part, err := writer.CreateFormFile("file", "audio.mp3")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, bytes.NewReader(audioData))
	if err != nil {
		return "", fmt.Errorf("failed to copy audio data: %w", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", a.url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.apiKey))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Text, nil
}

// Transcribe 根据 provider 调用相应的 ASR 服务
func (a *ASRClient) Transcribe(audioData []byte, language string) (string, error) {
	switch a.provider {
	case "openai":
		return a.TranscribeOpenAI(audioData, language)
	case "siliconflow":
		return a.TranscribeSiliconFlow(audioData, language)
	default:
		return "", fmt.Errorf("unsupported ASR provider: %s", a.provider)
	}
}

// TranscribeSiliconFlow 使用 SiliconFlow ASR 服务
func (a *ASRClient) TranscribeSiliconFlow(audioData []byte, language string) (string, error) {
	// SiliconFlow ASR 使用 multipart/form-data 格式
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// 添加模型参数
	_ = writer.WriteField("model", a.model)
	_ = writer.WriteField("language", language)

	// 上传音频文件（使用 .wav 后缀提示服务器格式）
	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, bytes.NewReader(audioData))
	if err != nil {
		return "", fmt.Errorf("failed to copy audio data: %w", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", a.url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.apiKey))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Text, nil
}
