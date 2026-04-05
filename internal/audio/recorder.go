package audio

import (
	"fmt"

	"github.com/gordonklaus/portaudio"
)

// Recorder 音频录制器
type Recorder struct {
	stream       *portaudio.Stream
	buffer       []float32
	sampleRate   int
	channels     int
	framesPerBuf int
}

// NewRecorder 创建新的录音机
func NewRecorder(sampleRate, channels, framesPerBuf int) (*Recorder, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize portaudio: %w", err)
	}

	buffer := make([]float32, framesPerBuf*channels)

	stream, err := portaudio.OpenDefaultStream(channels, 0, float64(sampleRate), framesPerBuf, buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to open portaudio stream: %w", err)
	}

	return &Recorder{
		stream:       stream,
		buffer:       buffer,
		sampleRate:   sampleRate,
		channels:     channels,
		framesPerBuf: framesPerBuf,
	}, nil
}

// Start 开始录音
func (r *Recorder) Start() error {
	return r.stream.Start()
}

// Stop 停止录音
func (r *Recorder) Stop() error {
	return r.stream.Stop()
}

// Close 关闭录音机
func (r *Recorder) Close() error {
	if r.stream != nil {
		_ = r.stream.Stop()
		err := r.stream.Close()
		portaudio.Terminate()
		return err
	}
	return nil
}

// Read 读取一帧音频数据
func (r *Recorder) Read() ([]float32, error) {
	err := r.stream.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	// 返回缓冲区的副本
	data := make([]float32, len(r.buffer))
	copy(data, r.buffer)
	return data, nil
}

// GetSampleRate 获取采样率
func (r *Recorder) GetSampleRate() int {
	return r.sampleRate
}

// GetFramesPerBuffer 获取每个缓冲区的帧数
func (r *Recorder) GetFramesPerBuffer() int {
	return r.framesPerBuf
}
