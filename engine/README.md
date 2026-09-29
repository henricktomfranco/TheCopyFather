# Embedded Engine Directory

Place the standalone `llama-server.exe` (from llama.cpp Windows release) in this directory.

- **Binary Name**: `llama-server.exe`
- **Hardware Mode**: CPU (`-ngl 0` / 0 GPU layers offloaded)
- **Threads**: Automatically detected, capped to configurable limit
- **Context Window**: 2048 - 4096 tokens
- **Thinking**: OFF
- **Streaming**: ON

The application automatically launches this engine on demand as a silent background child process and shuts it down cleanly when TheCopyFather exits. No external apps or server daemons are needed.
