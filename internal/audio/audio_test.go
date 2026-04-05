package audio

import (
	"testing"
)

// TestVADDetection 测试 VAD 语音检测
func TestVADDetection(t *testing.T) {
	vad := NewVAD(0.02, 500, 16000)

	// 测试无声帧（应该返回 false）
	silentFrame := make([]float32, 2048)
	for i := range silentFrame {
		silentFrame[i] = 0.001 // 非常小的值
	}

	isActive, _ := vad.DetectActivity(silentFrame, len(silentFrame))
	if isActive {
		t.Error("Expected silent frame to be inactive")
	}

	// 测试有声帧（应该返回 true）
	voiceFrame := make([]float32, 2048)
	for i := range voiceFrame {
		voiceFrame[i] = 0.1 // 较大的值
	}

	isActive, _ = vad.DetectActivity(voiceFrame, len(voiceFrame))
	if !isActive {
		t.Error("Expected voice frame to be active")
	}
}

// TestVADReset 测试 VAD 重置
func TestVADReset(t *testing.T) {
	vad := NewVAD(0.02, 500, 16000)

	// 设置为活跃状态
	voiceFrame := make([]float32, 2048)
	for i := range voiceFrame {
		voiceFrame[i] = 0.1
	}
	vad.DetectActivity(voiceFrame, len(voiceFrame))

	if !vad.IsActive() {
		t.Error("Expected VAD to be active")
	}

	// 重置
	vad.Reset()

	if vad.IsActive() {
		t.Error("Expected VAD to be inactive after reset")
	}
}

// TestAudioFileSplit 测试音频文件分割
func TestAudioFileSplit(t *testing.T) {
	// 创建测试数据（模拟音频）
	testData := make([]byte, 100000) // 100KB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	// 创建临时音频文件
	tmpFile := "/tmp/test_audio.bin"
	SaveToFile(tmpFile, string(testData))

	audioFile, err := NewAudioFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create AudioFile: %v", err)
	}

	// 分割测试
	chunks := audioFile.Split(25000) // 25KB 分割

	expectedChunks := 4 // 100KB / 25KB = 4
	if len(chunks) != expectedChunks {
		t.Errorf("Expected %d chunks, got %d", expectedChunks, len(chunks))
	}

	// 验证每个分割块的大小
	for i, chunk := range chunks {
		if len(chunk) > 25000 {
			t.Errorf("Chunk %d exceeds max size: %d", i, len(chunk))
		}
	}
}
