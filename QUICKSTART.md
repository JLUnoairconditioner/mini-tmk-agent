# Mini TMK Agent - 快速开始指南

## ⚡ 5分钟快速上手

### 1️⃣ 安装依赖

```bash
# 下载 Go 依赖
go mod download

# macOS 额外步骤（安装 PortAudio）
brew install portaudio

# Linux 额外步骤 (Ubuntu/Debian)
sudo apt-get install portaudio19-dev
```

### 2️⃣ 配置 API 密钥

```bash
# 复制配置模板
cp .env.example .env

# 编辑 .env，填入你的 API 密钥
# 获取渠道：
# - OpenAI: https://platform.openai.com/api-keys
# - DeepSeek: https://platform.deepseek.com/
# - 阿里云通义千问: https://dashscope.console.aliyun.com/
```

### 3️⃣ 编译程序

```bash
go build -o mini-tmk-agent ./cmd/mini-tmk-agent
```

### 4️⃣ 运行！

#### Stream 模式（实时翻译）

```bash
# 中文 → 英文
./mini-tmk-agent stream --source-lang zh --target-lang en

# 英文 → 中文（带详细日志）
./mini-tmk-agent stream --source-lang en --target-lang zh --verbose
```

**操作方法：**
- 说话会自动录音
- 程序会实时转录和翻译
- 按 Ctrl+C 停止

#### Transcript 模式（处理音频文件）

```bash
# 转录音频文件
./mini-tmk-agent transcript --file audio.mp3 --output result.txt

# 中文音频转英文文本
./mini-tmk-agent transcript \
  --file meeting.mp3 \
  --output meeting_en.txt \
  --source-lang zh \
  --target-lang en

# 仅转录不翻译
./mini-tmk-agent transcript \
  --file audio.mp3 \
  --output transcript.txt \
  --translate=false
```

---

## 📚 常见场景

### 场景 1：会议实时翻译

```bash
# 正在进行的会议是英文的，需要实时翻译成中文
./mini-tmk-agent stream --source-lang en --target-lang zh

# 监听麦克风输入，实时输出中文翻译
# 输出示例：
# [EN] The project timeline is 3 months
# [ZH] 项目时间轴是3个月
```

### 场景 2：讲座录音转录

```bash
# 有一个西班牙语讲座的录音，需要转录到英文
./mini-tmk-agent transcript \
  --file lecture.mp3 \
  --output lecture_english.txt \
  --source-lang es \
  --target-lang en \
  --verbose
```

### 场景 3：多语言对话

```bash
# 中日交流，需要双向翻译
# 监听中文
./mini-tmk-agent stream --source-lang zh --target-lang ja &

# 在另一个终端监听日文
./mini-tmk-agent stream --source-lang ja --target-lang zh
```

---

## 🔧 Makefile 快捷命令

```bash
# 构建
make build

# 只运行 Stream 模式
make run-stream

# 只运行 Transcript 模式
make run-transcript

# 格式化代码
make format

# 清理编译产物
make clean

# 显示所有命令
make help
```

---

## ⚠️ 常见问题

### Q: 麦克风没有声音
**A:** 
- 检查系统是否允许应用访问麦克风
- 确认默认输入设备设置正确
- 运行时加 `--verbose` 查看调试信息

### Q: API 返回 401 错误
**A:** 
- 检查 API_KEY 是否正确
- 检查 API_KEY 是否过期
- 检查 API_URL 是否正确

### Q: 转录结果不准确
**A:**
- 使用 `--verbose` 查看原始输出
- 检查音频质量（16kHz 采样率效果最佳）
- 尝试不同的 ASR 提供商

### Q: 翻译内容有问题
**A:**
- 检查源语言是否识别正确
- 尝试更强的模型 (如 gpt-4)
- 调整翻译 API 的 temperature 参数

---

## 📝 支持的语言

| 代码 | 名称 | 示例 |
|------|------|------|
| `zh` | 中文 | 你好，如何使用这个程序？ |
| `en` | 英文 | Hello, how to use this program? |
| `es` | 西班牙文 | Hola, ¿cómo usar este programa? |
| `ja` | 日文 | こんにちは、このプログラムの使い方は？ |

---

## 🎯 下一步

- 📖 阅读 [README.md](README.md) 了解更多细节
- 🔍 查看 [DEVELOPMENT.md](DEVELOPMENT.md) 学习架构和扩展
- 🐛 遇到问题？检查 [FAQ](README.md#常见问题排查)

---

**提示：** 首次运行时，程序会自动下载 Go 依赖并初始化 PortAudio。可能需要几秒时间。

祝你使用愉快！🚀
