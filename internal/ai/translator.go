package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// TranslationClient 翻译客户端
type TranslationClient struct {
	provider string
	apiKey   string
	url      string
	model    string
	client   *http.Client
}

// NewTranslationClient 创建翻译客户端
func NewTranslationClient(provider, apiKey, url, model string) *TranslationClient {
	return &TranslationClient{
		provider: provider,
		apiKey:   apiKey,
		url:      url,
		model:    model,
		client:   &http.Client{},
	}
}

// TranslateOpenAI 使用 OpenAI API 翻译
func (t *TranslationClient) TranslateOpenAI(text, sourceLang, targetLang string) (string, error) {
	prompt := fmt.Sprintf("Translate the following text from %s to %s. Only return the translated text without any explanation:\n\n%s",
		sourceLang, targetLang, text)

	payload := map[string]interface{}{
		"model": t.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3,
		"max_tokens":  2000,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", t.url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no translation result found")
}

// TranslateDeepSeek 使用 DeepSeek API 翻译
func (t *TranslationClient) TranslateDeepSeek(text, sourceLang, targetLang string) (string, error) {
	prompt := fmt.Sprintf("Translate the following text from %s to %s. Only return the translated text without any explanation:\n\n%s",
		sourceLang, targetLang, text)

	payload := map[string]interface{}{
		"model": t.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", t.url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no translation result found")
}

// Translate 根据 provider 调用相应的翻译服务
func (t *TranslationClient) Translate(text, sourceLang, targetLang string) (string, error) {
	switch t.provider {
	case "openai":
		return t.TranslateOpenAI(text, sourceLang, targetLang)
	case "deepseek":
		return t.TranslateDeepSeek(text, sourceLang, targetLang)
	case "qwen":
		// Qwen 通常兼容 OpenAI 的接口
		return t.TranslateOpenAI(text, sourceLang, targetLang)
	default:
		return "", fmt.Errorf("unsupported translation provider: %s", t.provider)
	}
}
