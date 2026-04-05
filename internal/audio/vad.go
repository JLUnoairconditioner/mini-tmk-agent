package audio

import (
	"math"
)

// VAD Voice Activity Detector - 语音活动检测器
type VAD struct {
	threshold           float32 // 能量阈值
	silenceDurationMs   int     // 静音持续时间阈值（毫秒）
	sampleRate          int
	silenceFrameCounter int
	isActive            bool
}

// NewVAD 创建新的 VAD 实例
func NewVAD(threshold float32, silenceDurationMs, sampleRate int) *VAD {
	return &VAD{
		threshold:           threshold,
		silenceDurationMs:   silenceDurationMs,
		sampleRate:          sampleRate,
		silenceFrameCounter: 0,
		isActive:            false,
	}
}

// DetectActivity 检测音频活动，返回是否有语音活动
func (v *VAD) DetectActivity(frame []float32, frameSize int) (bool, float64) {
	energy := v.calculateEnergy(frame)

	if energy > float64(v.threshold) {
		// 检测到语音
		v.silenceFrameCounter = 0
		v.isActive = true
		return true, energy
	}

	// 没有检测到语音，计算静音帧
	silenceDurationPerFrame := float64(v.silenceDurationMs) / 1000.0 / (float64(frameSize) / float64(v.sampleRate))
	v.silenceFrameCounter++

	if v.silenceFrameCounter > int(silenceDurationPerFrame) {
		// 检测到足够长的静音段，判定为语音结束
		v.isActive = false
		return false, energy
	}

	return v.isActive, energy
}

// calculateEnergy 计算音频帧的能量
func (v *VAD) calculateEnergy(frame []float32) float64 {
	var sum float64
	for _, sample := range frame {
		sum += float64(sample * sample)
	}
	avgEnergy := sum / float64(len(frame))
	return math.Sqrt(avgEnergy)
}

// IsActive 返回当前是否处于活跃状态
func (v *VAD) IsActive() bool {
	return v.isActive
}

// Reset 重置 VAD 状态
func (v *VAD) Reset() {
	v.silenceFrameCounter = 0
	v.isActive = false
}
