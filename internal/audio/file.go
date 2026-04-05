package audio

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// AudioFile 音频文件处理器
type AudioFile struct {
	filePath string
	file     *os.File
	data     []byte
}

// NewAudioFile 创建音频文件处理器
func NewAudioFile(filePath string) (*AudioFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to read audio file: %w", err)
	}

	_ = file.Close()

	return &AudioFile{
		filePath: filePath,
		data:     data,
	}, nil
}

// GetData 获取音频文件的字节数据
func (af *AudioFile) GetData() []byte {
	return af.data
}

// GetSize 获取音频文件大小（字节）
func (af *AudioFile) GetSize() int {
	return len(af.data)
}

// Split 将音频文件分割成多个块（用于处理大文件）
// 返回的是每个块的字节切片
func (af *AudioFile) Split(chunkSizeBytes int) [][]byte {
	var chunks [][]byte
	reader := bytes.NewReader(af.data)

	for {
		chunk := make([]byte, chunkSizeBytes)
		n, err := reader.Read(chunk)
		if err != nil && err != io.EOF {
			return chunks
		}

		if n > 0 {
			chunks = append(chunks, chunk[:n])
		}

		if err == io.EOF {
			break
		}
	}

	return chunks
}

// SaveToFile 保存数据到文件
func SaveToFile(filePath string, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}
