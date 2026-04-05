# Mini TMK Agent - 开发者指南

本文档为 Mini TMK Agent 的开发者提供详细的技术指南。

## 目录

1. [架构设计](#架构设计)
2. [代码结构](#代码结构)
3. [核心模块详解](#核心模块详解)
4. [扩展指南](#扩展指南)
5. [测试策略](#测试策略)
6. [性能优化](#性能优化)
7. [故障排查](#故障排查)

## 架构设计

### 全局架构

Mini TMK Agent 采用**生产者-消费者**模式和**管道处理**思想：

```
┌─────────────┐
│   CLI 入口  │ (cmd/mini-tmk-agent/main.go)
└──────┬──────┘
       │
       ▼
┌──────────────────┐
│  Cobra CLI 框架  │ (internal/cli/*)
├──────────────────┤
│ - Root Command   │
│ - Stream Command │
│ - Transcript Cmd │
└──────┬───────────┘
       │
       ▼
┌─────────────────────────────────────┐
│      2 种工作模式                    │
├─────────────────────────────────────┤
│  Stream Mode    │  Transcript Mode   │
│  ┌──────────┐   │  ┌──────────────┐  │
│  │ 麦克风   │   │  │ 读取音频文件  │  │
│  │ VAD      │   │  │ ASR API      │  │
│  │ ASR API  │   │  │ 翻译 API     │  │
│  │ 翻译 API │   │  │ 写入结果     │  │
│  └──────────┘   │  └──────────────┘  │
└─────────────────────────────────────┘
       │
       ▼
┌────────────────────┐
│   AI 服务接口      │ (internal/ai/*)
├────────────────────┤
│ - ASR 客户端       │
│ - 翻译客户端       │
└────────┬───────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌─────────┐ ┌──────────┐
│ OpenAI  │ │ DeepSeek │ 等外部 API
└─────────┘ └──────────┘
```

### 并发模型 (Stream Mode)

采用**多 Goroutine + Channel** 的并发设计：

```go
// 三个独立的 Goroutine，通过 Channel 通信

Goroutine 1: Audio Capture
    ┌──────────────────┐
    │ 读取麦克风数据   │
    │ VAD 检测语音片段 │
    │ 发送到 audioChan │
    └─────────┬────────┘
              │ []float32
              ▼
           audioChan

Goroutine 2: ASR Recognition
    ┌──────────────────┐
    │ 从 audioChan 读取│
    │ 调用 ASR API     │
    │ 发送到 textChan  │
    └─────────┬────────┘
              │ string
              ▼
           textChan

Goroutine 3: Translation
    ┌──────────────────┐
    │ 从 textChan 读取 │
    │ 调用翻译 API     │
    │ 输出到终端       │
    └──────────────────┘
```

**优势：**
- 解耦各个处理阶段
- 充分利用 I/O 等待时间
- 自动流控（Channel Buffer）

## 代码结构

### 核心文件对应表

| 文件路径 | 职责 | 关键函数/类型 |
|---------|------|-------------|
| `cmd/mini-tmk-agent/main.go` | 程序入口 | `main()` |
| `internal/cli/root.go` | CLI 根命令 | `NewRootCmd()` |
| `internal/cli/stream.go` | Stream 子命令 | `NewStreamCmd()`, `runStreamCmd()` |
| `internal/cli/transcript.go` | Transcript 子命令 | `NewTranscriptCmd()`, `runTranscriptCmd()` |
| `internal/audio/recorder.go` | 麦克风录音 | `Recorder`, `NewRecorder()` |
| `internal/audio/vad.go` | 语音检测 | `VAD`, `DetectActivity()` |
| `internal/audio/file.go` | 文件读取 | `AudioFile`, `SaveToFile()` |
| `internal/ai/asr.go` | 语音识别 API | `ASRClient`, `Transcribe()` |
| `internal/ai/translator.go` | 翻译 API | `TranslationClient`, `Translate()` |
| `internal/config/config.go` | 配置管理 | `Config`, `LoadConfig()` |
| `pkg/logger/logger.go` | 日志输出 | `PrintSourceText()`, `PrintTargetText()` |

## 核心模块详解

### 1. Audio 模块 (`internal/audio/`)

#### Recorder (recorder.go)

使用 **PortAudio** 库进行跨平台音频采集：

```go
// 初始化
recorder, err := audio.NewRecorder(16000, 1, 2048)

// 启动录音
recorder.Start()

// 循环读取
for {
    frame, err := recorder.Read()  // 返回一帧音频数据
    // 处理 frame
}

// 关闭
recorder.Close()
```

**流程：**
1. `Initialize()` - 初始化 PortAudio
2. `OpenDefaultStream()` - 打开默认输入设备
3. `Read()` - 阻塞读取，直到获得一帧数据
4. `Close()` - 释放资源

#### VAD (vad.go)

**语音活动检测** - 判断何时有人说话：

```go
vad := audio.NewVAD(0.02, 500, 16000)  // 阈值、静音时长、采样率

// 每帧检测
isActive := vad.DetectActivity(frame, len(frame))

if !isActive {
    // 检测到足够长的静音 -> 发送音频到 ASR
}
```

**算法原理：**
- 计算每帧的**能量** (RMS): $E = \sqrt{\frac{1}{N}\sum_{i=0}^{N-1}x_i^2}$
- 与阈值比较判断是否有语音
- 统计连续静音帧数，超过阈值则判定语音结束

#### AudioFile (file.go)

处理本地音频文件：

```go
audioFile, err := audio.NewAudioFile("audio.mp3")
data := audioFile.GetData()           // 获取字节数据

// 分割大文件
chunks := audioFile.Split(25*1024*1024)  // 25MB 分割
```

### 2. AI 模块 (`internal/ai/`)

#### ASRClient (asr.go)

调用语音识别 API：

```go
asrClient := ai.NewASRClient("openai", apiKey, url)

// 支持多个提供商
text, err := asrClient.Transcribe(audioBytes, "zh")

// 内部根据 provider 分发
switch provider {
    case "openai":
        return a.TranscribeOpenAI(...)
    case "siliconflow":
        return a.TranscribeSiliconFlow(...)
}
```

**OpenAI Whisper 集成：**
- Multipart form 上传音频文件
- 返回 JSON：`{"text": "转录结果"}`
- 自动处理语言参数

#### TranslationClient (translator.go)

调用翻译 API：

```go
translationClient := ai.NewTranslationClient("openai", apiKey, url)

// 支持多个 LLM 提供商
translatedText, err := translationClient.Translate(text, "zh", "en")
```

**Chat API 集成示例 (OpenAI)：**

```json
{
  "model": "gpt-3.5-turbo",
  "messages": [
    {
      "role": "user",
      "content": "Translate to English: 你好"
    }
  ],
  "temperature": 0.3
}
```

### 3. CLI 模块 (`internal/cli/`)

使用 **Cobra** 框架构建 CLI：

```go
// 应用结构
root
├── stream
│   ├── --source-lang (zh, en, es, ja)
│   ├── --target-lang (zh, en, es, ja)
│   └── --verbose
└── transcript
    ├── --file (必需)
    ├── --output (必需)
    ├── --source-lang
    ├── --target-lang
    ├── --translate
    └── --verbose
```

**Cobra 三个关键概念：**

```go
// 1. Command - 命令本身
cmd := &cobra.Command{
    Use:   "stream",
    Short: "Description",
    RunE:  runStreamCmd,  // 执行函数
}

// 2. Flag - 命令行参数
cmd.Flags().String("source-lang", "zh", "Help text")

// 3. Args - 命令行参数值
sourceLang, _ := cmd.Flags().GetString("source-lang")
```

## 扩展指南

### 添加新的 ASR 提供商

**步骤：**

1. **在 `internal/ai/asr.go` 中添加新方法：**

```go
func (a *ASRClient) TranscribeNewProvider(audioData []byte, language string) (string, error) {
    // 1. 构建请求
    payload := buildPayload(audioData, language)
    
    // 2. 发送 HTTP 请求
    resp, err := a.client.Do(req)
    if err != nil {
        return "", fmt.Errorf("request failed: %w", err)
    }
    
    // 3. 解析响应
    result := parseResponse(resp)
    
    return result.Text, nil
}
```

2. **在 `Transcribe` 方法中添加 case：**

```go
func (a *ASRClient) Transcribe(audioData []byte, language string) (string, error) {
    switch a.provider {
    case "openai":
        return a.TranscribeOpenAI(audioData, language)
    case "newprovider":  // 新增
        return a.TranscribeNewProvider(audioData, language)
    default:
        return "", fmt.Errorf("unsupported provider: %s", a.provider)
    }
}
```

3. **更新 `.env.example`：**

```bash
ASR_PROVIDER=newprovider
ASR_API_KEY=key...
ASR_URL=https://api.newprovider.com/asr
```

### 添加新的翻译提供商

**步骤：** 类似 ASR，编辑 `internal/ai/translator.go`

```go
func (t *TranslationClient) TranslateNewProvider(text, sourceLang, targetLang string) (string, error) {
    // 实现思路同上
}

// 在 Translate 中添加路由
case "newprovider":
    return t.TranslateNewProvider(text, sourceLang, targetLang)
```

### 添加新语言支持

**步骤：**

1. **编辑 `internal/cli/stream.go` 和 `transcript.go`：**

```go
func isValidLanguage(lang string) bool {
    validLanguages := map[string]bool{
        "zh": true,  // 中文
        "en": true,  // 英文
        "es": true,  // 西班牙文
        "ja": true,  // 日文
        "fr": true,  // 法文 (新增)
        "de": true,  // 德文 (新增)
    }
    return validLanguages[lang]
}

func getLanguageName(lang string) string {
    names := map[string]string{
        "zh": "Chinese",
        "en": "English",
        "es": "Spanish",
        "ja": "Japanese",
        "fr": "French",      // 新增
        "de": "German",      // 新增
    }
    if name, ok := names[lang]; ok {
        return name
    }
    return lang
}
```

### 添加新的输出格式

**步骤：** 编辑 `internal/cli/transcript.go` 中的 `generateOutput` 函数

```go
func generateOutput(...) string {
    switch outputFormat {
    case "text":
        return generateText(...)
    case "json":  // 新增
        return generateJSON(...)
    case "markdown":  // 新增
        return generateMarkdown(...)
    }
}
```

## 测试策略

### 单元测试

创建 `internal/audio/recorder_test.go`：

```go
package audio

import "testing"

func TestNewRecorder(t *testing.T) {
    recorder, err := NewRecorder(16000, 1, 2048)
    if err != nil {
        t.Fatalf("NewRecorder failed: %v", err)
    }
    defer recorder.Close()
    
    if recorder.GetSampleRate() != 16000 {
        t.Errorf("expected 16000, got %d", recorder.GetSampleRate())
    }
}
```

### 集成测试

测试 Stream 模式的完整流程（需要配置 API）

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/audio

# 详细输出
go test -v ./...

# 显示覆盖率
go test -cover ./...
```

## 性能优化

### Stream Mode 优化要点

| 优化点 | 策略 | 效果 |
|-------|------|------|
| **延迟** | 减小 VADSilenceDuration | 更快检测说话结束 |
| **吞吐** | 增大 FramesPerBuffer | 减少函数调用开销 |
| **准确性** | 调整 VADThreshold | 避免误判 |
| **内存** | 及时释放 Channel buffer | 防止内存溢出 |

### Transcript Mode 优化要点

| 优化点 | 策略 | 备注 |
|-------|------|------|
| **大文件** | 分割处理（<25MB） | OpenAI API 限制 |
| **并发** | 多 Goroutine 并发处理 | 使用 sync.WaitGroup |
| **缓存** | 缓存 API 响应 | 减少重复请求 |
| **压缩** | 压缩音频数据 | 节省传输带宽 |

### Profiling (性能分析)

```bash
# CPU 性能分析
go run -cpuprofile=cpu.prof ./cmd/mini-tmk-agent stream ...

# 内存分析
go run -memprofile=mem.prof ./cmd/mini-tmk-agent stream ...

# 查看分析结果
go tool pprof cpu.prof
```

## 故障排查

### 常见错误

#### 1. PortAudio 初始化失败

```
Error: failed to initialize portaudio: ...
```

**解决：**
```bash
# macOS
brew install portaudio

# Linux
sudo apt-get install portaudio19-dev

# Windows: 下载官方二进制
```

#### 2. API 认证失败

```
Error: API error (status 401): Unauthorized
```

**检查清单：**
- [ ] API_KEY 是否正确
- [ ] API_URL 是否正确
- [ ] API_KEY 是否过期

#### 3. 网络超时

```
Error: context deadline exceeded
```

**解决：**
- 增加 HTTP 超时时间
- 检查网络连接
- 使用代理

#### 4. 内存泄漏

**排查：**
```bash
# 生成内存分析
go run -memprofile=mem.prof cmd/mini-tmk-agent/main.go

# 分析
go tool pprof mem.prof
pprof> top -cum
```

### 调试技巧

**1. 启用详细日志：**
```bash
./mini-tmk-agent stream --verbose
```

**2. 打印中间变量：**
```go
logger.PrintInfo(fmt.Sprintf("Debug: %v", variable))
```

**3. 使用 Delve 调试器：**
```bash
# 安装
go install github.com/go-delve/delve/cmd/dlv@latest

# 启动调试
dlv debug ./cmd/mini-tmk-agent
> b main.main
> c
```

---

**最后更新：** 2026-04-04
**维护者：** Mini TMK Team
