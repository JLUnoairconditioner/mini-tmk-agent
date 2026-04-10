# Mini TMK Agent - 音频转录和翻译工具

一个强大的 Go CLI 工具，将音频采集、语音识别 (ASR) 和机器翻译集成为一个无缝管道。

## 概述

Mini TMK Agent (Transcription & Machine Translation Kit) 提供两种工作模式：

> ⚠️ 重要提醒：
> - Windows 用户在 `stream` 模式下需要本地可用的 PortAudio 动态链接库（如 `portaudio.dll`）。
> - 仓库中未包含 Windows 平台的音频库二进制文件，请自行安装 PortAudio 或通过 MinGW/GCC 编译环境准备依赖。
> - 如果运行过程中遇到网络代理问题，请清空终端代理环境变量后重试：
>   - PowerShell:
>     ```powershell
>     $env:HTTP_PROXY=""
>     $env:HTTPS_PROXY=""
>     $env:ALL_PROXY=""
>     ```
>   - Bash / WSL:
>     ```bash
>     unset HTTP_PROXY HTTPS_PROXY ALL_PROXY
>     ```

### 模式一：流式同传 (Stream Mode)
实时监听麦克风，自动转录并翻译你的语音。支持实时上次翻译，适合会议、讲座等场景。

```bash
mini-tmk-agent stream --source-lang zh --target-lang en
```

**特性：**
- 🎤 实时麦克风监听
- 🎯 Voice Activity Detection (VAD) 自动断句
- ⚡ 低延迟流式处理
- 🌍 支持多语言 (中、英、西、日)
- 🎨 彩色终端输出

### 模式二：文件转录 (Transcript Mode)
处理本地音频文件，生成转录和翻译文本。

```bash
mini-tmk-agent transcript --file audio.mp3 --output result.txt --source-lang zh --target-lang en
```

**特性：**
- 📁 支持多种音频格式 (MP3, WAV, PCM, M4A, FLAC)
- 📝 生成格式化的转录文本
- 🔄 自动翻译并整合结果
- 📊 详细的处理报告

## 技术栈

| 组件 | 技术选型 |
|------|--------|
| **CLI框架** | [spf13/cobra](https://github.com/spf13/cobra) |
| **音频采集** | [gordonklaus/portaudio](https://github.com/gordonklaus/portaudio) |
| **语音识别** | OpenAI Whisper API / SiliconFlow ASR |
| **语言翻译** | OpenAI / DeepSeek / Qwen LLM |
| **UI/UX** | [pterm](https://github.com/pterm/pterm) (彩色输出) |
| **环境管理** | [godotenv](https://github.com/joho/godotenv) |

## 项目结构

```
mini-tmk-agent/
├── cmd/
│   └── mini-tmk-agent/
│       └── main.go              # 程序入口
├── internal/
│   ├── audio/                   # 音频模块
│   │   ├── recorder.go          # 麦克风录音
│   │   ├── vad.go               # 语音活动检测
│   │   └── file.go              # 文件读取
│   ├── ai/                      # AI 模块
│   │   ├── asr.go               # 语音转文字 API
│   │   └── translator.go        # 翻译 API
│   ├── cli/                     # CLI 模块
│   │   ├── root.go              # 根命令
│   │   ├── stream.go            # Stream 子命令
│   │   └── transcript.go        # Transcript 子命令
│   └── config/                  # 配置模块
│       └── config.go            # 配置加载
├── pkg/
│   └── logger/                  # 日志工具
│       └── logger.go
├── go.mod
├── go.sum
├── .env.example                 # 环境变量示例
└── README.md                    # 本文件
```

## 快速开始

### 1. 环境要求

- Go 1.21+
- Git
- 麦克风 (用于 stream 模式)
- API 密钥 (OpenAI / DeepSeek / Qwen等)

### 2. 依赖安装

#### Windows/macOS/Linux 通用

```bash
# 克隆项目
git clone <repository-url>
cd mini-tmk-agent

# 下载依赖
go mod download
```

#### macOS 额外依赖 (PortAudio)

```bash
brew install portaudio
```

#### Linux 额外依赖 (Ubuntu/Debian)

```bash
sudo apt-get install portaudio19-dev
```

#### Windows 额外依赖

下载并安装 [PortAudio Binaries](http://www.portaudio.com/download.html)

### 3. 配置 API 密钥

创建 `.env` 文件：

```bash
cp .env.example .env
```

编辑 `.env` 并填入你的 API 密钥：

Windows 环境运行须知：
由于本项目 stream 模式依赖底层的 PortAudio C语言库，考虑到跨平台兼容性与 Git 规范，代码仓库中未包含 .dll 二进制文件。Windows 开发者拉取代码后，请确保本地已配置 MinGW/GCC 编译环境，或在根目录自行补充相关的音频动态链接库（如 portaudio.dll）后再执行 go run。

```bash
# ASR 服务配置
ASR_PROVIDER=openai              # openai 或 siliconflow
ASR_API_KEY=sk-xxxxxxxxxxxxx
ASR_URL=https://api.openai.com/v1/audio/transcriptions

# 翻译服务配置
TRANSLATION_PROVIDER=openai      # openai, deepseek, qwen
TRANSLATION_API_KEY=sk-xxxxxxxxxxxxx
TRANSLATION_URL=https://api.openai.com/v1/chat/completions

# 音频配置
VAD_ENABLED=true
```

### 4. 编译

```bash
# 编译可执行文件
go build -o mini-tmk-agent ./cmd/mini-tmk-agent

# 或直接运行
go run ./cmd/mini-tmk-agent stream --source-lang zh --target-lang en

> 如果在运行 `go run` 时遇到端口冲突或网络代理导致的连接问题，请先清空当前终端的代理设置：
>
> ```powershell
$env:HTTP_PROXY=""
$env:HTTPS_PROXY=""
$env:ALL_PROXY=""
> ```
>
> 然后重新运行命令。

### 5. 使用示例

#### Stream 模式

```bash
# 中文转英文
./mini-tmk-agent stream --source-lang zh --target-lang en

# 英文转中文（详细输出）
./mini-tmk-agent stream --source-lang en --target-lang zh --verbose

# 西班牙文转英文
./mini-tmk-agent stream --source-lang es --target-lang en
```

#### Transcript 模式

```bash
# 基础用法
go run ./cmd/mini-tmk-agent transcript --file meeting.mp3 --output result.txt

go run ./cmd/mini-tmk-agent transcript --file test.mp3 --output result.txt --source-lang en --target-lang zh 
# 指定源和目标语言
./mini-tmk-agent transcript \
  --file audio.mp3 \
  --output transcription.txt \
  --source-lang zh \
  --target-lang en

# 仅转录，不翻译
./mini-tmk-agent transcript \
  --file audio.wav \
  --output transcription.txt \
  --translate=false

# 详细输出
./mini-tmk-agent transcript \
  --file audio.mp3 \
  --output result.txt \
  --verbose
```

## 支持的语言

| 代码 | 语言 | 示例 |
|------|------|------|
| `zh` | 中文 | 你好 |
| `en` | 英文 | Hello |
| `es` | 西班牙文 | Hola |
| `ja` | 日文 | こんにちは |

## 工作流程

### Stream Mode 工作流

```
┌──────────────┐     ┌────────┐     ┌─────────┐     ┌────────────┐
│ 麦克风输入   │────▶│ VAD    │────▶│ ASR     │────▶│ 翻译       │
│ Microphone   │     │ 检测   │     │ 识别    │     │ Translation│
└──────────────┘     └────────┘     └─────────┘     └────────────┘
                          │               │               │
                          ▼               ▼               ▼
                     [断句检测]      [实时识别]      [即时翻译]
                     [Silent Pause] [Live Text]     [Live Output]
```

**各阶段详解：**

1. **Voice Activity Detection (VAD)**
   - 实时监听麦克风数据流
   - 检测用户何时开始/停止说话
   - 利用能量阈值和静音时长判断断句点

2. **Speech Recognition (ASR)**
   - 接收 VAD 分割的音频块
   - 调用 Whisper API 进行识别
   - 返回转录后的源语言文本

3. **Translation**
   - 接收转录的源语言文本
   - 调用 LLM API 进行翻译
   - 实时输出目标语言文本

### Transcript Mode 工作流

```
┌──────────────────┐     ┌─────────┐     ┌────────────┐     ┌─────────┐
│ 读取音频文件    │────▶│ ASR     │────▶│ 翻译       │────▶│ 写入    │
│ Read Audio File  │     │ 识别    │     │ Translation│     │ 结果    │
└──────────────────┘     └─────────┘     └────────────┘     └─────────┘
```

## API 配置指南

### OpenAI Whisper (推荐 ASR)

```bash
ASR_PROVIDER=openai
ASR_API_KEY=sk-xxxxxx  # 获取：https://platform.openai.com/api-keys
ASR_URL=https://api.openai.com/v1/audio/transcriptions
```

### DeepSeek (推荐翻译)

```bash
TRANSLATION_PROVIDER=deepseek
TRANSLATION_API_KEY=sk-xxxxxx  # 获取：https://platform.deepseek.com/
TRANSLATION_URL=https://api.deepseek.com/chat/completions
```

### 阿里云通义千问 (Qwen)

```bash
TRANSLATION_PROVIDER=qwen
TRANSLATION_API_KEY=sk-xxxxxx
TRANSLATION_URL=https://api.qwen.aliyun.com/v1/chat/completions
```

## 限制和已知问题

### 文件大小限制
- OpenAI Whisper API: 最大 25MB
- 超大文件会自动分割处理

### 语言支持
- 当前支持 4 种语言 (中、英、西、日)
- 可以通过修改 `isValidLanguage()` 函数扩展

### 网络要求
- 需要稳定的互联网连接
- 某些地区可能需要代理配置

## 性能建议

### Stream Mode 优化

| 参数 | 推荐值 | 说明 |
|------|-------|------|
| `SampleRate` | 16000 Hz | 标准采样率，平衡质量和性能 |
| `VADThreshold` | 0.02 | 能量阈值，根据环境调整 |
| `VADSilenceDuration` | 500 ms | 静音时长，通常 400-800ms |
| `FramesPerBuffer` | 2048 | 每次读取帧数，越小延迟越低 |

### Transcript Mode 优化

在处理大文件时，考虑：
- 分割成较小的块（<25MB）
- 并发处理多个文件
- 使用缓存减少 API 调用

## 开发指南

### 添加新的 ASR 提供商

编辑 `internal/ai/asr.go`：

```go
func (a *ASRClient) TranscribeNewProvider(audioData []byte, language string) (string, error) {
    // 实现新的 ASR 逻辑
}

// 在 Transcribe 方法中添加 case
case "newprovider":
    return a.TranscribeNewProvider(audioData, language)
```

### 添加新的翻译提供商

编辑 `internal/ai/translator.go`：

```go
func (t *TranslationClient) TranslateNewProvider(text, sourceLang, targetLang string) (string, error) {
    // 实现新的翻译逻辑
}

// 在 Translate 方法中添加 case
case "newprovider":
    return t.TranslateNewProvider(text, sourceLang, targetLang)
```

### 添加新的语言支持

1. 编辑 `internal/cli/stream.go` 和 `transcript.go`
2. 在 `isValidLanguage()` 中添加语言代码
3. 在 `getLanguageName()` 中添加显示名称

## 调试和日志

启用详细日志输出：

```bash
# Stream 模式
./mini-tmk-agent stream --source-lang zh --target-lang en --verbose

# Transcript 模式
./mini-tmk-agent transcript --file audio.mp3 --output result.txt --verbose
```

## 常见问题排查

### Q: 麦克风无法识别
**A:** 检查 PortAudio 安装和系统权限
```bash
# 确认 PortAudio 库已安装
# macOS
brew list | grep portaudio
# Linux
dpkg -l | grep portaudio
```

### Q: ASR API 超时
**A:** 检查网络连接，考虑使用本地 ASR 模型或增加超时时间

### Q: 翻译质量不理想
**A:** 调整 LLM 的温度参数 (temperature)，或更换为更强的模型

### Q: Stream 模式延迟高
**A:** 
- 减小 `VADSilenceDuration` 值
- 增大 `FramesPerBuffer` 值
- 检查网络带宽

## 贡献指南

欢迎提交 Pull Request！

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

MIT License - 详见 LICENSE 文件

## 联系方式和支持

- 📧 Email: xuzz0322@gmail.com
- 🐛 Issues: GitHub Issues
- 💬 讨论: GitHub Discussions

## 致谢

- [spf13/cobra](https://github.com/spf13/cobra) - CLI 框架
- [gordonklaus/portaudio](https://github.com/gordonklaus/portaudio) - 音频处理
- [pterm/pterm](https://github.com/pterm/pterm) - 终端美化

---

**备注：** 这是一个演示项目，API 调用会产生费用。请合理使用 API，了解相关定价信息。
