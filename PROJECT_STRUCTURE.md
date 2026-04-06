## 📋 项目完整结构总览

```
mini-tmk-agent/
│
├── 📄 【配置与文档】
│   ├── go.mod                    # Go 模块定义
│   ├── go.sum                    # 依赖锁定
│   ├── .env.example              # 环境变量模板
│   └── .gitignore                # Git 忽略规则
│
├── 📖 【文档】
│   ├── README.md                 # 项目总览（推荐首先阅读）
│   ├── QUICKSTART.md             # 快速开始指南（5分钟上手）
│   ├── DEVELOPMENT.md            # 开发者指南（架构、扩展、测试）
│   ├── ARCHITECTURE.md           # 架构详解（并发模型、设计模式）
│   └── PROJECT_STRUCTURE.md      # 本文件
│
├── 🚀 【可执行入口】
│   └── cmd/mini-tmk-agent/
│       └── main.go               # 程序入口点
│
├── 📦 【核心代码 - internal/】
│   │
│   ├── cli/                      # CLI 命令模块
│   │   ├── root.go               # 根命令定义，子命令注册
│   │   ├── stream.go             # Stream 模式实现（实时翻译）
│   │   └── transcript.go         # Transcript 模式实现（文件处理）
│   │
│   ├── audio/                    # 音频处理模块
│   │   ├── recorder.go           # 麦克风录音（基于 PortAudio）
│   │   ├── vad.go                # 语音活动检测（Voice Activity Detection）
│   │   ├── file.go               # 音频文件读取与处理
│   │   └── audio_test.go         # 单元测试
│   │
│   ├── ai/                       # AI 服务接口模块
│   │   ├── asr.go                # 语音识别 API 客户端
│   │   │                         # ├─ TranscribeOpenAI()
│   │   │                         # ├─ TranscribeSiliconFlow()
│   │   │                         # └─ Transcribe() (路由)
│   │   │
│   │   └── translator.go         # 语言翻译 API 客户端
│   │                             # ├─ TranslateOpenAI()
│   │                             # ├─ TranslateDeepSeek()
│   │                             # └─ Translate() (路由)
│   │
│   └── config/                   # 配置管理模块
│       └── config.go             # 从环境变量加载配置
│                                 # ├─ ASR 配置
│                                 # ├─ 翻译配置
│                                 # ├─ 音频配置
│                                 # └─ VAD 配置
│
└── 🛠️ 【工具库 - pkg/】
    └── logger/
        └── logger.go             # 日志输出工具
                                 # ├─ PrintSourceText() - 绿色
                                 # ├─ PrintTargetText() - 蓝色
                                 # ├─ PrintSuccess()
                                 # ├─ PrintError()
                                 # └─ PrintWarning()
```

---

## 📍 关键文件概览

### 🎯 快速参考

| 文件 | 用途 | 关键类型 |
|-----|------|---------|
| `cmd/main.go` | 程序入口 | 调用 CLI 框架 |
| `internal/cli/root.go` | CLI 框架初始化 | `NewRootCmd()` |
| `internal/cli/stream.go` | Stream 模式 | 生产者-消费者模式 |
| `internal/cli/transcript.go` | Transcript 模式 | 顺序处理 |
| `internal/audio/recorder.go` | 麦克风输入 | `Recorder` 类型 |
| `internal/audio/vad.go` | 语音检测 | `VAD` 类型 |
| `internal/audio/file.go` | 文件I/O | `AudioFile` 类型 |
| `internal/ai/asr.go` | 语音识别 | `ASRClient` 类型 |
| `internal/ai/translator.go` | 翻译服务 | `TranslationClient` 类型 |
| `internal/config/config.go` | 配置加载 | `Config` 类型 |

---

## 🔄 执行流程

### Stream Mode 流程

```
1. CLI 入口
   ├─ cmd/mini-tmk-agent/main.go (main函数)
   └─ → internal/cli/root.go (NewRootCmd)
            → internal/cli/stream.go (runStreamCmd)

2. 初始化阶段
   ├─ config.LoadConfig()        # 加载 .env 配置
   ├─ audio.NewRecorder()        # 初始化麦克风
   ├─ ai.NewASRClient()          # 创建 ASR 客户端
   ├─ ai.NewTranslationClient()  # 创建翻译客户端
   └─ audio.NewVAD()             # 创建 VAD 检测器

3. 并发处理 (3 个 Goroutine)
   ├─ G1: 录音 + VAD
   │   └─ recorder.Read() → vad.DetectActivity() → audioChan
   │
   ├─ G2: ASR 识别
   │   └─ audioChan → asrClient.Transcribe() → textChan
   │
   └─ G3: 翻译输出
       └─ textChan → translationClient.Translate() → logger.Print()

4. 信号处理
   └─ Ctrl+C → 关闭 recorder → 关闭通道 → 等待 Goroutine 完成
```

### Transcript Mode 流程

```
1. CLI 入口
   └─ internal/cli/transcript.go (runTranscriptCmd)

2. 验证阶段
   ├─ 检查文件存在
   ├─ 验证语言代码
   ├─ 验证输出目录
   └─ 检查 API 密钥

3. 处理阶段
   ├─ audio.NewAudioFile()       # 读取音频文件
   ├─ asrClient.Transcribe()     # 调用 ASR API
   ├─ translationClient.Translate() # 调用翻译 API
   └─ generateOutput()           # 格式化输出

4. 保存阶段
   └─ audio.SaveToFile()         # 保存到输出文件
```

---

## 🧩 模块依赖关系

```
┌────────────────────────────────────────────┐
│ 外部依赖库                                   │
├────────────────────────────────────────────┤
│ • gordonklaus/portaudio - 音频采集          │
│ • spf13/cobra - CLI 框架                   │
│ • pterm/pterm - 彩色输出                   │
│ • joho/godotenv - 环境变量加载             │
│ • 标准库: net/http, encoding/json 等      │
└────────────────────────────────────────────┘
                    ▲
                    │ import
                    │
┌────────────────────────────────────────────┐
│ pkg/ (可复用工具)                           │
├────────────────────────────────────────────┤
│ • logger.go - 彩色输出工具                 │
└────────────────────────────────────────────┘
                    ▲
                    │ import
                    │
┌────────────────────────────────────────────┐
│ internal/ (内部模块)                        │
├────────────────────────────────────────────┤
│ • audio/ - 音频处理                        │
│ • ai/ - AI 服务调用                        │
│ • config/ - 配置管理                       │
│ • cli/ - 命令行界面                        │
└────────────────────────────────────────────┘
                    ▲
                    │ import
                    │
┌────────────────────────────────────────────┐
│ cmd/mini-tmk-agent/ (可执行入口)           │
├────────────────────────────────────────────┤
│ • main.go - 程序入口                       │
└────────────────────────────────────────────┘
```

---

## 💾 核心数据结构

### Config (internal/config/config.go)
```go
type Config struct {
    // ASR 配置
    ASRProvider  string    // openai, siliconflow
    ASRAPIKey    string
    ASRURL       string

    // 翻译配置
    TranslationProvider string  // openai, deepseek, qwen
    TranslationAPIKey   string
    TranslationURL      string

    // 音频配置
    SampleRate     int  // 16000
    BitDepth       int  // 16
    NumChannels    int  // 1
    FramesPerBuffer int // 2048

    // VAD 配置
    VADEnabled       bool     // true
    VADThreshold     float32  // 0.02
    VADSilenceDuration int    // 500ms
}
```

### Recorder (internal/audio/recorder.go)
```go
type Recorder struct {
    stream       *portaudio.Stream
    buffer       []float32
    sampleRate   int
    channels     int
    framesPerBuf int
}
```

### VAD (internal/audio/vad.go)
```go
type VAD struct {
    threshold           float32
    silenceDurationMs   int
    sampleRate          int
    silenceFrameCounter int
    isActive            bool
}
```

### ASRClient (internal/ai/asr.go)
```go
type ASRClient struct {
    provider string
    apiKey   string
    url      string
    client   *http.Client
}
```

### TranslationClient (internal/ai/translator.go)
```go
type TranslationClient struct {
    provider string
    apiKey   string
    url      string
    client   *http.Client
}
```

---

## 🔌 API 集成点

### ASR API 调用
```
┌─────────────────────┐
│ 用户说话/音频文件    │
└──────────┬──────────┘
           │
           ▼
    ┌─────────────────┐
    │ WAV/PCM 格式    │
    └────────┬────────┘
             │ Multipart form
             ▼
    ┌───────────────────────────┐
    │ OpenAI Whisper API        │
    │ POST /audio/transcriptions │
    └────────┬──────────────────┘
             │
      ┌──────▼──────┐
      ▼             ▼
    成功           失败
  ┌──────┐      ┌─────────┐
  │JSON  │      │404/401  │
  └──────┘      └─────────┘
    │
    ▼
  {"text": "转录内容"}
```

### Translation API 调用
```
┌────────────────────┐
│ ASR 转录的文本      │
└────────┬───────────┘
         │ "你好"
         ▼
┌──────────────────────────┐
│ Chat API (OpenAI/等)     │
│ POST /chat/completions   │
└────────┬─────────────────┘
         │
    ┌────▼─────┐
    ▼          ▼
  成功        失败
┌──────┐   ┌─────────┐
│JSON  │   │401/429  │
└──────┘   └─────────┘
   │
   ▼
{
  "choices": [{
    "message": {
      "content": "Hello"
    }
  }]
}
```

---

## 📊 线程/Goroutine 配置

### Stream Mode
- **Goroutine 1** (Audio Capture): 优先级 Normal，CPU密集
- **Goroutine 2** (ASR): 优先级 Normal，I/O 密集（等待 API）
- **Goroutine 3** (Translation): 优先级 Normal，I/O 密集（等待 API）
- **Main Goroutine**: 阻塞等待信号，协调关闭

### Channel 配置
- `audioChan`: buffer = 10 (音频块)
- `textChan`: buffer = 10 (文本块)

---

## 🚀 编译和运行

### 编译
```bash
# 生成可执行文件
go build -o mini-tmk-agent ./cmd/mini-tmk-agent
```

### 运行
```bash
# Stream 模式
./mini-tmk-agent stream --source-lang zh --target-lang en

# Transcript 模式
./mini-tmk-agent transcript --file audio.mp3 --output result.txt
```

---

## 📝 配置文件

### .env 文件结构
```bash
# ASR 配置
ASR_PROVIDER=openai
ASR_API_KEY=sk-...
ASR_URL=https://api.openai.com/v1/audio/transcriptions

# 翻译配置
TRANSLATION_PROVIDER=openai
TRANSLATION_API_KEY=sk-...
TRANSLATION_URL=https://api.openai.com/v1/chat/completions

# 可选：音频配置
VAD_ENABLED=true
```

---

## ✅ 项目完整性检查

- ✅ CLI 框架集成 (Cobra)
- ✅ 两种工作模式 (Stream / Transcript)
- ✅ 音频输入处理 (Recorder)
- ✅ 语音活动检测 (VAD)
- ✅ ASR 集成 (OpenAI / SiliconFlow)
- ✅ 翻译集成 (OpenAI / DeepSeek / Qwen)
- ✅ 配置管理 (环境变量)
- ✅ 日志输出 (彩色终端)
- ✅ 错误处理
- ✅ 并发控制 (Goroutine + Channel)
- ✅ 资源管理 (defer)
- ✅ 测试框架
- ✅ 完整文档

---

## 🎓 推荐阅读顺序

1. **README.md** - 了解项目概况
2. **QUICKSTART.md** - 5分钟快速上手
3. **ARCHITECTURE.md** - 理解架构设计
4. **DEVELOPMENT.md** - 开发和扩展
5. **代码** - 按照上述流程图阅读源码

---

**生成日期：** 2026-04-04  
**项目状态：** ✅ 完整  
**推荐 Go 版本：** 1.21+
