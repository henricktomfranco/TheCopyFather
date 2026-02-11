# The Copyfather

> *"I'm gonna make you an offer you can't refuse... better writing"*

An AI-powered text rewriting and analysis tool that runs seamlessly in the background. Like a ghostwriter, but with more *family connections*.

![The Copyfather](https://img.shields.io/badge/AI-Powered-blue) ![Windows](https://img.shields.io/badge/Platform-Windows-green) ![Ollama](https://img.shields.io/badge/Ollama-Compatible-orange)

## 🎬 What's This?

**The Copyfather** is your personal AI writing assistant that helps you rewrite, analyze, and improve text using local AI models via Ollama. Named after the famous line from *The Godfather*, this app makes you an offer you can't refuse: better writing, instantly.

## ✨ Features

### 🔄 **Rewrite Modes**
- **Grammar & Spelling** - Professional editing with error correction
- **Paraphrase** - Fresh wording while keeping meaning
- **Standard** - Balanced, natural rewrite
- **Formal** - Professional tone for business documents
- **Casual** - Conversational, friendly style
- **Creative** - Expressive and vivid language
- **Short** - Concise and punchy
- **Expand** - More detail and depth

### 📊 **Analysis Modes**
- **TL;DR Summary** - 2-3 sentence condensed version
- **Key Points** - Bullet list of main points (beautifully formatted!)
- **Key Insights** - Important facts and arguments

### 🖥️ **Smart Windows Integration**
- **Global Hotkey** (default: `Ctrl+Shift+R`) - Works in any app
- **Auto-Paste** - Replace text without manual copy/paste
- **Position Near Cursor** - Popup appears where you're working
- **System Tray** - Runs quietly in the background
- **"Don't Ask Again"** - Remember your paste preferences

### 🎨 **Visual Highlights**
- **Bold Important Words** - AI highlights key terms automatically
- **Dark Theme** - Easy on the eyes
- **Email Detection** - Special formatting for professional emails
- **Bullet Lists** - Properly rendered for readability

## 🚀 Installation

### Prerequisites
1. **Go 1.21+** - [Download](https://go.dev/dl/)
2. **Node.js 18+** - [Download](https://nodejs.org/)
3. **Wails CLI** - Run: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
4. **Ollama** - [Download](https://ollama.ai/)

### Build from Source

```bash
# Clone or navigate to the project
cd C:\Go\paraChange

# Install Go dependencies
go mod tidy

# Install frontend dependencies
cd frontend
npm install
cd ..

# Build the application
wails build -platform windows/amd64
```

The built executable will be at: `build/bin/thecopyfather.exe`

## 🎯 First Run

1. **Start Ollama**: Make sure Ollama is running (`ollama serve` or system tray)
2. **Run the app**: Double-click `thecopyfather.exe`
3. **Configure**:
   - Right-click system tray icon → Settings
   - Set Ollama server (default: http://localhost:11434)
   - Select model (e.g., gemma3:1b, llama2, mistral)
   - Click "Test Connection"
   - Configure Auto-Paste behavior (Ask/Always/Never)
   - Save settings

4. **Use it**:
   - Select text in any application
   - Press `Ctrl+Shift+R`
   - Choose your rewrite/analysis style
   - Click "Replace Selection"
   - The improved text automatically replaces your selection!

## 🎮 Usage

### Global Hotkey
- **Default**: `Ctrl+Shift+R`
- **Customizable** in Settings

### Auto-Paste Modes
- **Ask Every Time** - Shows confirmation dialog with "Don't ask again" checkbox
- **Always Paste** - Automatically pastes without asking
- **Never Paste** - Copies to clipboard only, you paste manually

### Popup Position
- **Near Cursor** (default) - Appears next to where you're typing
- **Screen Center** - Traditional centered window

## ⚙️ Configuration

Settings stored in: `%APPDATA%\TheCopyfather\config.json`

```json
{
  "server_url": "http://localhost:11434",
  "model": "gemma3:1b",
  "api_key": "",
  "default_style": "grammar",
  "auto_start": true,
  "hotkey": "ctrl+shift+r",
  "monitor_clipboard": false,
  "auto_paste_mode": "ask",
  "popup_position_mode": "cursor"
}
```

## 🔧 Ollama Setup

```bash
# 1. Install Ollama (https://ollama.ai)

# 2. Pull a model
ollama pull gemma3:1b
# or
ollama pull llama2

# 3. Start the server
ollama serve

# 4. Test it's running
curl http://localhost:11434/api/tags
```

## 🐛 Troubleshooting

### Hotkey Not Working
- Some applications (elevated/admin) may block global hotkeys
- Try running The Copyfather as administrator
- Change the hotkey combination in Settings

### Clipboard Monitoring Issues
- Some apps use private clipboard APIs
- Use the hotkey method as a reliable alternative
- Enable "Monitor clipboard" in Settings

### Auto-Paste Not Working
- Ensure the target app has focus before triggering
- Try increasing the delay in Settings (if available)
- Some secure applications block automated paste operations

### Build Failures
```bash
# Clean build
go clean -cache
rm -rf frontend/node_modules
rm -rf frontend/dist
cd frontend && npm install && npm run build
cd ..
wails build -platform windows/amd64
```

## 🎨 Why "The Copyfather"?

> *"Leave the gun. Take the cannoli. And fix that grammar."*

This app is your **consigliere** for writing - always there in the background, ready to help you make your text an offer it can't refuse. Whether you're drafting emails, writing reports, or just need to paraphrase something quickly, The Copyfather has your back.

Like Don Corleone, this app:
- ✅ Commands respect (from your text)
- ✅ Solves problems (grammar, style, clarity)
- ✅ Works behind the scenes (system tray)
- ✅ Makes you look good (professional writing)

## 📁 Project Structure

```
thecopyfather/
├── main.go                     # Application entry point
├── wails.json                  # Wails configuration
├── go.mod                      # Go dependencies
├── README.md                   # This file
│
├── internal/
│   ├── config/
│   │   └── config.go          # Settings management
│   ├── ollama/
│   │   └── client.go          # Ollama API client
│   ├── rewriter/
│   │   └── rewriter.go        # Text processing & prompts
│   └── windows/
│       ├── hotkey.go          # Global hotkey registration
│       ├── clipboard.go       # Clipboard operations
│       ├── tray.go            # System tray integration
│       ├── window.go          # Window utilities
│       └── autostart.go       # Auto-start with Windows
│
└── frontend/
    ├── index.html
    ├── package.json
    └── src/
        ├── main.tsx
        ├── App.tsx
        ├── components/
        │   ├── Popup.tsx
        │   └── Settings.tsx
        └── styles/
```

## 📝 License

MIT License - Feel free to use, modify, and distribute!

## 🙏 Credits

- Built with [Wails](https://wails.io/) - Build desktop apps with Go + React
- Powered by [Ollama](https://ollama.ai/) - Run LLMs locally
- Inspired by *The Godfather* (1972) - "I'm gonna make him an offer he can't refuse"

---

> *"A writer writes. But a Copyfather? He makes it better."*

**Made with ❤️ and a little bit of family business.**
