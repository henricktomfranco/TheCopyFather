# The Copyfather - Feature Analysis & Implementation Roadmap

> Generated from a full codebase analysis. Existing features are complete. New features are categorized by effort.

---

## Already Implemented (✅ Complete)

| # | Feature | Description |
|---|---------|-------------|
| 1 | **Rewrite Modes** (8 styles) | Grammar, Paraphrase, Standard, Formal, Casual, Creative, Short, Expand |
| 2 | **Analysis Modes** (3 styles) | TL;DR Summary, Key Points, Key Insights |
| 3 | **Global Hotkey** | Ctrl+Shift+R (configurable), Windows RegisterHotKey API |
| 4 | **Auto-Paste** | Ask / Always / Never modes with "Don't ask again" |
| 5 | **System Tray** | Background running with Settings/Exit menu |
| 6 | **Smart Positioning** | Popup near cursor or screen center |
| 7 | **Clipboard Monitoring** | Optional polling-based clipboard watching |
| 8 | **Ollama Integration** | Health check, model listing, version detection, legacy fallback |
| 9 | **Auto-start with Windows** | Desktop entry / registry for autostart |
| 10 | **Dark Theme UI** | Modern glassmorphism dark theme |
| 11 | **Smart Text Type Detection** | Auto-detects: Email, Chat, Code, List, Normal text |
| 12 | **Type-Specific Prompts** | Different prompts per text type (email vs code vs chat vs list) |
| 13 | **Custom Prompts** | Per-style, per-text-type custom AI prompts |
| 14 | **Rich HTML Clipboard** | Bold formatting, email structure, list formatting in clipboard |
| 15 | **Formatting Toggle** | Switch between bold-formatted and plain text |
| 16 | **Mini Mode** | Compact floating widget with quick rewrite |
| 17 | **Undo/Redo History** | Ctrl+Z / Ctrl+Y within the popup, up to 20 entries |
| 18 | **Confidence Score** | UI display of rewrite confidence |
| 19 | **Config Encryption** | AES-GCM encryption for API keys |
| 20 | **Exponential Backoff Retry** | 3 retries with backoff on Ollama failures |
| 21 | **XML Sanitization** | Input escaping to prevent prompt injection |
| 22 | **Response Cleaning** | Strips XML tags, thinking blocks, conversational filler from AI output |
| 23 | **Logger** | Structured logging (Info, Error, Warning, Debug) |
| 24 | **Multi-monitor Support** | Screen bounds checking for window positioning |
| 25 | **Welcome Screen** | First-run privacy/permissions notice |
| 26 | **Streaming Responses** | Stream AI output token-by-token via Ollama `/api/generate` with `stream: true` |
| 27 | **History / Session Log** | SQLite DB implementation with save, list, delete, favorite, and toggle operations |
| 28 | **Keyboard Navigation** | Full keyboard support with hotkey triggers and navigation |
| 29 | **Diff View** | ComputeDiff() method for comparing original and rewritten text |
| 30 | **Tone Analyzer** | ToneAnalysis with detectTone() for analyzing text tone |
| 31 | **Wails Frontend** | Complete React + TypeScript frontend with all components |
| 32 | **Model Management** | GetAvailableModels(), TestConnection(), version checking |
| 33 | **Style Info System** | GetStyleInfo() for rewrite and analysis styles |
| 34 | **Text Type Info** | GetTextTypeInfo() with normalized scores |
| 35 | **Error Handling** | Comprehensive error logging and user feedback |
| 36 | **Window Management** | Top-most, positioning, minimize/restore, always-on-top |
| 37 | **Clipboard Rich Text** | Markdown to HTML conversion, rich text formatting |
| 38 | **Hotkey Management** | Register, parse, validate, simulate copy/paste |
| 39 | **Configuration System** | Load, save, default prompts, custom prompts management |
| 40 | **Crypto System** | AES-GCM encryption/decryption for sensitive data |

---

## Features to Implement (🆕 Proposed)

### Priority 1 - High Impact / Low Effort

| # | Feature | Description | Effort |
|---|---------|-------------|--------|
| 1 | **Multiple AI Provider Support** | Add OpenAI, Anthropic, and other cloud LLM APIs alongside Ollama. Abstract client interface. | High |
| 2 | **Export Formats** | Export rewritten text as: PDF, DOCX, Markdown file, plain text. | Medium |
| 3 | **Accessibility (a11y)** | Screen reader support (ARIA labels), high-contrast theme, font size settings. | Medium |
| 4 | **Drag & Drop Text** | Accept dragged text files or selections directly on the popup. | Low |
| 5 | **Text-to-Speech** | Read original or rewritten text aloud using system TTS. | Low |

### Priority 2 - Medium Impact

| # | Feature | Description | Effort |
|---|---------|-------------|--------|
| 6 | **Bulk Rewrite** | Rewrite multiple selected text entries in batch. | Medium |
| 7 | **User Styles / Presets** | Save custom style combinations (e.g., "Business Email" preset with Formal + Short + Email type). | Medium |
| 8 | **Collaboration / Sync** | Sync custom prompts and settings across devices via GitHub Gist or Dropbox. | Medium |
| 9 | **Quick Actions** | Define custom hotkeys for specific actions (e.g., Ctrl+Shift+1 = Formal, Ctrl+Shift+2 = Short). | Medium |
| 10 | **Spelling & Grammar Check** | Integrate with LanguageTool or similar for pre-rewrite spellcheck. | Medium |

### Priority 3 - Advanced / Nice-to-Have

| # | Feature | Description | Effort |
|---|---------|-------------|--------|
| 11 | **Cross-Platform Support** | Port to macOS and Linux. Requires replacing Windows-specific APIs (hotkey, clipboard, tray) with platform abstractions. | High |
| 12 | **Plugin System** | Allow third-party plugins for new styles, custom AI backends, or output formats. | High |
| 13 | **Team / Shared Prompts** | Sync custom prompts across a team via a shared config file or server. | High |
| 14 | **AI Model Fine-tuning** | Allow users to fine-tune their own models for specific use cases. | High |
| 15 | **Voice Input** | Speak text to be rewritten instead of typing/pasting. | Medium |

---

## Technical Debt / Improvements

| # | Item | Description |
|---|------|-------------|
| T1 | **Unit Tests** | Only one test file exists (`prompt_test.go`). Need coverage for config, ollama client, rewriter, clipboard, hotkeys. |
| T2 | **Integration Tests** | End-to-end tests for the full rewrite flow with a mock Ollama server. |
| T3 | **Error Telemetry** | Optional crash reporting to help debugging (opt-in). |
| T4 | **Config Migration** | Version config schema and auto-migrate old configs when format changes. |
| T5 | **Hardcoded AES Key** | `crypto.go` uses a fixed key `1234567890123456` — should use OS keychain or derive from machine ID. |
| T6 | **Go Deprecations** | `GenerateRewrites` and `GenerateAnalysis` are marked deprecated but still present. Clean up. |
| T7 | **Frontend Error Boundaries** | React error boundaries around Popup, Settings, MiniMode to prevent white screens. |
| T8 | **CSS Modularization** | Large inline styles in Popup.tsx (~400 lines). Extract to CSS modules. |
| T9 | **Type Safety** | Heavy use of `@ts-ignore` for Wails bindings. Add proper type generation. |
| T10 | **CI/CD Pipeline** | GitHub Actions for build, test, lint, and release artifacts. |

---

## Summary

- **25 features already implemented** and working.
- **30 new features proposed** across 3 priority tiers.
- **10 technical improvements** identified.
- **Quick wins** (low effort, high impact): Diff View, Keyboard Navigation, Drag & Drop, TTS.
- **Biggest missing features**: Streaming responses, multi-provider support, session history, cross-platform.