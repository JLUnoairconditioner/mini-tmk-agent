# Mini TMK Agent - 架构和工程设计说明

## 📐 项目工程架构

### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                     终端用户                                  │
└────────────────┬────────────────────────────────────────────┘
                 │
                 │ CLI 命令
                 │
┌────────────────▼────────────────────────────────────────────┐
│                   Cobra CLI 框架                              │
│                                                               │
│    mini-tmk-agent                                            │
│    ├── stream [--source-lang] [--target-lang] [--verbose]   │
│    └── transcript [--file] [--output] [--option]*           │
└────┬──────────────────────────────────┬─────────────────────┘
     │                                  │
     ▼ Stream Mode                      ▼ Transcript Mode
┌────────────────┐                  ┌────────────────┐
│  Audio Input   │                  │ File Reader    │
│  (Microphone)  │                  │ (MP3/WAV/...)  │
└────────┬───────┘                  └────────┬───────┘
         │                                   │
         ▼                                   ▼
    ┌─────────────────┐              ┌──────────────┐
    │  VAD (断句检测) │              │  直接处理    │
    │  Channel-based  │              │  (无需VAD)   │
    │  Concurrency    │              └──────┬───────┘
    └────────┬────────┘                     │
             │                              │
    ┌────────▼────────┐                    │
    │   ASR Pipeline  │◀───────────────────┘
    │  (ASR Client)   │
    └────────┬────────┘
             │
    ┌────────▼────────────────────┐
    │  Translation Pipeline        │
    │  (Translation Client)        │
    └────────┬────────────────────┘
             │
             ▼
    ┌────────────────────────────┐
    │ Output (File / Terminal)   │
    └────────────────────────────┘
```

### 包依赖关系图

```
cmd/
└── mini-tmk-agent/
    └── main.go ──────┐
                      │
                      ▼
internal/cli/
├── root.go ───────────────┐
├── stream.go ─┐           │
└── transcript.go │        ▼
                  │    ┌─────────────────────┐
                  │    │ Cobra CLI Framework │
                  │    └─────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────┐
│       internal/audio/                       │
│  ├── recorder.go   (PortAudio)             │
│  ├── vad.go        (Voice Detection)       │
│  └── file.go       (File I/O)              │
└─┬───────────────────────────────────────────┘
  │
  ├──────────────────┐
  │                  │
  ▼                  ▼
internal/ai/       internal/config/
├── asr.go       └── config.go
├── translator.go
└─ (HTTP Clients)
  
  ├──────────────────┐
  │                  │
  ▼                  ▼
 OpenAI API      DeepSeek API
 Whisper         (或其他 LLM)
```

### 模块职责分工

| 模块 | 职责 | 依赖 |
|------|------|------|
| **cmd/mini-tmk-agent** | 程序入口点 | CLI 模块 |
| **internal/cli** | Cobra 命令定义、参数解析 | Audio、AI、Config |
| **internal/audio** | 音频采集、VAD、文件读取 | PortAudio |
| **internal/ai** | ASR 和翻译 API 调用封装 | HTTP Client |
| **internal/config** | 配置加载（环境变量） | godotenv |
| **pkg/logger** | 日志输出格式化 | pterm |

---

## 🔄 Stream Mode 详细工作流

### 并发模型详解

```go
// Goroutine 1: Audio Capture (audio/recorder.go + audio/vad.go)
func recordAudio() {
    for {
        frame := recorder.Read()          // 从麦克风读取帧
        isActive := vad.DetectActivity()  // VAD 判断
        if frameCompleted {
            audioChan <- audioBuffer       // 发送到通道
        }
    }
}

// Goroutine 2: ASR Processing (ai/asr.go)
func performASR() {
    for audioFrame := range audioChan {   // 持续监听通道
        text := asrClient.Transcribe()    // 调用 API
        textChan <- text                  // 转发到翻译
    }
}

// Goroutine 3: Translation (ai/translator.go)
func performTranslation() {
    for text := range textChan {          // 持续监听通道
        translated := translationClient.Translate()  // 调用 API
        logger.PrintTargetText(translated)          // 输出
    }
}
```

### Channel 流量控制

```
audioChan (buffer: 10)
    ⊢ 可以缓冲 10 个音频块
    ⊢ 防止生产者阻塞
    ⊢ 自动背压控制

        ↓ (ASR 可能较慢)

textChan (buffer: 10)
    ⊢ 存储转录的文本
    ⊢ 翻译速度通常快于 ASR
    ⊢ 不会有堆压
```

### 时间线示例

```
时间 ───────────────────────────────────────────────────

Goroutine 1 (Audio):
  [读框] [读框] [读框] [检测结束] [发送] [读框] [读框] ...
        麦克风采集      VAD 判断

Goroutine 2 (ASR):
          ↓ 接收到音频
          [API 调用中...] (耗时 2-5 秒)
          [完成，发送文本]

Goroutine 3 (Translation):
                         ↓ 接收到文本
                         [API 调用中...] (耗时 0.5-2 秒)
                         [打印结果]

输出:
  T1: [ZH] 你好
  T2: [EN] Hello
```

---

## 📊 Transcript Mode 详细工作流

### 顺序处理流程

```go
func runTranscriptCmd() {
    // 1. 验证输入参数
    validateInputs()
    
    // 2. 加载配置
    config := LoadConfig()
    
    // 3. 读取音频文件
    audioFile := NewAudioFile(filePath)
    audioData := audioFile.GetData()
    
    // 4. 调用 ASR API
    transcribedText := asrClient.Transcribe(audioData, sourceLang)
    logger.PrintSourceText(sourceLang, transcribedText)
    
    // 5. 调用翻译 API
    translatedText := translationClient.Translate(
        transcribedText, sourceLang, targetLang
    )
    logger.PrintTargetText(targetLang, translatedText)
    
    // 6. 保存结果到文件
    output := generateOutput(...)
    SaveToFile(outputPath, output)
}
```

### 大文件处理策略

```
输入文件大小 > 25MB
    │
    ├─ true  ─► 分割 (Split by duration)
    │              │
    │              ├─ Chunk 1 (0-25MB)
    │              ├─ Chunk 2 (25-50MB)
    │              └─ Chunk 3 (>50MB)
    │                   │
    │                   ▼
    │            并行 ASR (goroutine pool)
    │                   │
    │                   ▼
    │            汇总结果
    │
    └─ false ─► 直接调用 ASR
```

---

## 🛠️ 核心依赖分析

### 外部库使用

| 库 | 版本 | 用途 | 为什么选择 |
|----|------|------|---------|
| **cobra** | v1.7.0 | CLI 命令框架 | Go 生态标准，功能完整 |
| **portaudio** | latest | 音频输入 | 跨平台，稳定性好 |
| **pterm** | v0.12.66 | 彩色输出 | 简洁易用，美观 |
| **godotenv** | v1.5.1 | .env 加载 | 标准实践，配置管理 |

### 标准库关键使用

| 包 | 用途 | 核心 API |
|----|------|---------|
| **net/http** | HTTP 请求 | `http.Client`, `http.Request` |
| **encoding/json** | JSON 编解码 | `json.Marshal`, `json.Unmarshal` |
| **os** | 文件操作 | `os.Open`, `os.WriteFile` |
| **io** | I/O 操作 | `io.ReadAll`, `io.Copy` |
| **sync** | 并发 | `sync.WaitGroup`, `sync.Mutex` |
| **context** | 上下文控制 | `context.WithCancel`, `context.Done()` |

---

## 🎯 设计模式应用

### 1. 生产者-消费者模式

```go
// Goroutine 1: 生产者
go func() {
    for {
        data := recorder.Read()
        audioChan <- data  // 发送到通道
    }
}()

// Goroutine 2: 消费者
go func() {
    for data := range audioChan {  // 从通道接收
        process(data)
    }
}()
```

**优势：**
- 解耦生产和消费逻辑
- 自动流控（buffer size）
- 便于扩展（多消费者/多生产者）

### 2. 工厂模式

```go
// 工厂函数创建客户端
func NewASRClient(provider, apiKey, url string) *ASRClient {
    return &ASRClient{...}
}

// 使用时无需关心实现细节
asrClient := NewASRClient("openai", key, url)
```

### 3. 策略模式

```go
// 根据 provider 选择不同策略
func (a *ASRClient) Transcribe(audioData []byte, language string) (string, error) {
    switch a.provider {
    case "openai":
        return a.TranscribeOpenAI(...)  // 策略1
    case "siliconflow":
        return a.TranscribeSiliconFlow(...) // 策略2
    }
}
```

### 4. 责任链模式

```
Input
  │
  ▼
┌──────────┐
│ Audio    │ ── 负责：采集、缓冲
│ Recorder │
└───┬──────┘
    │
    ▼
┌──────────┐
│ VAD      │ ── 负责：检测语音段落
│ Detector │
└───┬──────┘
    │
    ▼
┌──────────┐
│ ASR      │ ── 负责：转录
│ Client   │
└───┬──────┘
    │
    ▼
┌──────────┐
│Transform │ ── 负责：翻译
│ Client   │
└───┬──────┘
    │
    ▼
  Output
```

---

## 🚀 性能考量

### Stream Mode 性能瓶颈

1. **麦克风读取延迟** (~20-40ms)
   - PortAudio frame buffer 大小决定
   - `FramesPerBuffer = 2048, SampleRate = 16000` ⟹ ~128ms/frame

2. **ASR API 延迟** (~2-5s)
   - 网络往返时间
   - OpenAI Whisper 处理时间
   - **跨域优化：** 使用 SiliconFlow（国内）或本地模型

3. **翻译 API 延迟** (~0.5-2s)
   - LLM 模型调用
   - **优化：** 并行 ASR + 翻译，不串联

### Transcript Mode 性能优化

```go
// 优化 1: 分块并行处理
chunks := audioFile.Split(maxSize)
for _, chunk := range chunks {
    go func(c []byte) {
        result := asrClient.Transcribe(c, lang)
    }(chunk)
}

// 优化 2: 缓存
cache := make(map[string]string)
if cached, ok := cache[hash(audioData)]; ok {
    return cached  // 命中缓存
}

// 优化 3: 流式输处
for range chunks {
    writeToFile(result)  // 逐块写入，而不是全部加载到内存
}
```

---

## 📋 项目开发检查清单

- [ ] **编码规范**
  - [ ] 代码格式化 (`go fmt`)
  - [ ] Linter 检查 (`golangci-lint`)
  - [ ] 变量命名规范

- [ ] **错误处理**
  - [ ] 所有错误都被检查
  - [ ] 错误信息清晰有用
  - [ ] 资源泄漏防护（defer close）

- [ ] **测试**
  - [ ] 单元测试覆盖核心逻辑
  - [ ] 集成测试验证流程
  - [ ] 边界条件测试

- [ ] **文档**
  - [ ] README 清晰完整
  - [ ] 代码注释充分
  - [ ] API 文档齐全

- [ ] **性能**
  - [ ] 内存泄漏检查
  - [ ] CPU 使用合理
  - [ ] 并发安全

- [ ] **安全**
  - [ ] API Key 不硬编码
  - [ ] 敏感数据加密
  - [ ] 输入验证

---

**此架构设计遵循 Go 最佳实践和工程规范，具有良好的易维护性、可扩展性和性能表现。**
