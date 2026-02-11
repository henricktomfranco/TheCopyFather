# The Copyfather - Build Guide

> *"I'm gonna make you an offer you can't refuse... better writing"*

## 🎬 About This Project

**The Copyfather** is an AI-powered text rewriting and analysis application for Windows. Like a ghostwriter with "family connections," this app helps you improve your writing using local AI models via Ollama.

### Features
- **Rewrite Text** in multiple styles (Grammar, Formal, Casual, Creative, Short, Expand)
- **Analyze Text** with TL;DR summaries, bullet points, and key insights
- **Auto-Paste** - Replace selected text automatically (with confirmation dialog)
- **Smart Positioning** - Popup appears near your cursor
- **Global Hotkey** - Works from any application (default: Ctrl+Shift+R)
- **System Tray** - Runs quietly in the background

---

## 🚀 Quick Start

### Prerequisites
1. **Go** (1.21+) - https://go.dev/dl/
2. **Node.js** (18+) - https://nodejs.org/
3. **Wails CLI** - Run: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
4. **Ollama** - https://ollama.ai/

### Build Steps

```bash
# Step 1: Navigate to project directory
cd C:\Go\paraChange

# Step 2: Download Go dependencies
go mod tidy

# Step 3: Install frontend dependencies
cd frontend
npm install
cd ..

# Step 4: Build the application
wails build -platform windows/amd64
```

The built executable will be at: `build/bin/thecopyfather.exe`

### First Run
1. **Start Ollama**: Make sure Ollama is running (`ollama serve` or check system tray)
2. **Run the app**: Double-click `thecopyfather.exe`
3. **Configure**: 
   - Right-click system tray icon → Settings
   - Set Ollama server (default: http://localhost:11434)
   - Select model (e.g., gemma3:1b, llama2, mistral)
   - Click "Test Connection"
   - Configure Auto-Paste behavior (Ask/Always/Never)
   - Set Popup Position (Near cursor/Screen center)
   - Save settings
4. **Use it**:
   - Select text in any app
   - Press `Ctrl+Shift+R`
   - Choose rewrite or analyze mode
   - Select your preferred style
   - Click "Replace Selection"
   - Text is automatically improved and pasted!

---

## 📝 Configuration

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

---

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

---

## 🐛 Troubleshooting

### LSP Errors (Expected)
The LSP errors showing "undefined: wails" and import errors are **normal** before building. They will resolve after running `go mod tidy`.

### Hotkey Not Working
- Some applications (like elevated/admin apps) may block global hotkeys
- Try running The Copyfather as administrator
- Change the hotkey combination in settings

### Auto-Paste Not Working
- Ensure the target application has focus
- Try increasing the auto-paste delay in settings
- Some secure applications block automated paste
- Check Auto-Paste Mode setting (Ask/Always/Never)

### Clipboard Monitoring Issues
- Some applications use private clipboard APIs
- Use the hotkey method as a reliable alternative
- Ensure "Monitor clipboard" is enabled in settings

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

---

## 📦 Key Implementation Details

### Backend (Go)
- **Wails v2**: Native Windows app with React frontend
- **System Tray**: `github.com/getlantern/systray`
- **Hotkeys**: Windows RegisterHotKey API
- **Clipboard**: Windows clipboard APIs
- **Concurrent Rewrites**: Goroutines for parallel processing
- **Settings**: JSON file in APPDATA

### Frontend (React + TypeScript)
- **Components**: Popup, Settings, Welcome
- **Styling**: CSS with dark theme
- **Event System**: Wails events for frontend/backend communication
- **Type Safety**: TypeScript interfaces for all data

### Rewrite Prompts
Each style has a specialized prompt that tells the AI to:
- Rewrite the text appropriately
- **Bold important words** using `**markdown**`
- Maintain proper formatting
- Only return the rewritten text

### Analysis Prompts
- **summarize**: 2-3 sentence TL;DR
- **bullets**: 3-5 key bullet points
- **insights**: Key facts and arguments

---

## 🎯 Next Steps

1. **Build the project** following the steps above
2. **Install Ollama** and pull a model
3. **Run thecopyfather.exe**
4. **Configure** your Ollama settings
5. **Start rewriting text!**

---

## 🎨 The Name

> *"Leave the gun. Take the cannoli. And fix that grammar."*

**The Copyfather** is your consigliere for writing - always there in the background, ready to help you make your text an offer it can't refuse. Whether you're drafting emails, writing reports, or just need to paraphrase something quickly, The Copyfather has your back.

Like Don Corleone, this app:
- ✅ Commands respect (from your text)
- ✅ Solves problems (grammar, style, clarity)
- ✅ Works behind the scenes (system tray)
- ✅ Makes you look good (professional writing)

---

## 📞 Support

If you encounter issues:
1. Check README.md for detailed documentation
2. Verify Ollama is running: `curl http://localhost:11434/api/tags`
3. Check Windows Event Viewer for crash logs
4. Review the application code - everything is well-commented

---

## 🎉 Success!

You now have **The Copyfather** - a fully functional AI writing assistant that:
- Runs silently in the background
- Responds to global hotkeys
- Integrates with Ollama LLM
- Provides multiple writing styles
- Has a modern, dark-themed UI
- Auto-pastes improved text
- Can auto-start with Windows

**Enjoy better writing!** 🎬✨

> *"A writer writes. But a Copyfather? He makes it better."*
