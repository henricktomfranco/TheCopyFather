package engine

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Status represents the status of the embedded llama.cpp runtime
type Status struct {
	IsRunning     bool   `json:"is_running"`
	ModelLoaded   string `json:"model_loaded"`
	ModelFound    bool   `json:"model_found"`
	ModelPath     string `json:"model_path"`
	BinaryFound   bool   `json:"binary_found"`
	BinaryPath    string `json:"binary_path"`
	Port          int    `json:"port"`
	Threads       int    `json:"threads"`
	MaxThreads    int    `json:"max_threads"`
	ContextSize   int    `json:"context_size"`
	Hardware      string `json:"hardware"`
	ThinkingOff   bool   `json:"thinking_off"`
	StreamingOn   bool   `json:"streaming_on"`
	Error         string `json:"error,omitempty"`
}

// Config defines settings for the embedded engine
type Config struct {
	BinaryPath      string
	ModelPath       string
	Port            int
	ContextSize     int  // 2048 - 4096
	CPUThreads      int  // 0 = auto-detect
	MaxCPUThreads   int  // configurable limit
	DisableThinking bool // Thinking: OFF
	Streaming       bool // Streaming: ON
}

// Engine manages the local llama-server child process
type Engine struct {
	cfg        Config
	cmd        *exec.Cmd
	mu         sync.Mutex
	isRunning  bool
	actualPort int
	threads    int
	lastError  string
}

// EnsureDirectories ensures that the engine and models directories exist in AppData and project dir
func EnsureDirectories() {
	appData := os.Getenv("APPDATA")
	if appData != "" {
		_ = os.MkdirAll(filepath.Join(appData, "TheCopyfather", "engine"), 0755)
		_ = os.MkdirAll(filepath.Join(appData, "TheCopyfather", "models"), 0755)
	}
	_ = os.MkdirAll("engine", 0755)
	_ = os.MkdirAll("models", 0755)
}

// New creates a new embedded Engine manager
func New(cfg Config) *Engine {
	EnsureDirectories()

	if cfg.Port <= 0 {
		cfg.Port = 8085
	}
	if cfg.ContextSize <= 0 {
		cfg.ContextSize = 4096
	}
	if cfg.MaxCPUThreads <= 0 {
		numCPU := runtime.NumCPU()
		if numCPU > 6 {
			cfg.MaxCPUThreads = 6
		} else {
			cfg.MaxCPUThreads = numCPU
		}
	}
	return &Engine{
		cfg: cfg,
	}
}

// UpdateConfig updates the engine configuration
func (e *Engine) UpdateConfig(cfg Config) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.cfg = cfg
}

// DetectThreads calculates the thread count to use based on CPU and limits
func DetectThreads(cpuThreads, maxThreads int) int {
	numCPU := runtime.NumCPU()
	if cpuThreads > 0 {
		if maxThreads > 0 && cpuThreads > maxThreads {
			return maxThreads
		}
		return cpuThreads
	}
	// Auto-detect with configurable limit
	threads := numCPU
	if threads > 4 {
		threads = threads - 2 // leave UI / system threads responsive
	} else if threads > 2 {
		threads = threads - 1
	}
	if maxThreads > 0 && threads > maxThreads {
		threads = maxThreads
	}
	if threads < 1 {
		threads = 1
	}
	return threads
}

// GetDefaultEngineDir returns the standard %APPDATA%\TheCopyfather\engine directory
func GetDefaultEngineDir() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = os.Getenv("USERPROFILE")
		if appData == "" {
			appData = "."
		}
	}
	return filepath.Join(appData, "TheCopyfather", "engine")
}

// GetDefaultModelsDir returns the standard %APPDATA%\TheCopyfather\models directory
func GetDefaultModelsDir() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = os.Getenv("USERPROFILE")
		if appData == "" {
			appData = "."
		}
	}
	return filepath.Join(appData, "TheCopyfather", "models")
}

// GetDefaultBinaryPath returns the default path to llama-server.exe in AppData
func GetDefaultBinaryPath() string {
	return filepath.Join(GetDefaultEngineDir(), "llama-server.exe")
}

// GetDefaultModelPath returns the default path to SmolLM2-360M GGUF in AppData
func GetDefaultModelPath() string {
	return filepath.Join(GetDefaultModelsDir(), "smollm2-360m-instruct-q4_k_m.gguf")
}

// FindBinary searches for llama-server.exe with AppData as the primary standard location
func FindBinary(customPath string) (string, bool) {
	candidates := []string{}
	if customPath != "" {
		candidates = append(candidates, customPath)
	}

	// 1. AppData directory (PRIMARY for installed/distributed app)
	appData := os.Getenv("APPDATA")
	if appData != "" {
		candidates = append(candidates,
			filepath.Join(appData, "TheCopyfather", "engine", "llama-server.exe"),
			filepath.Join(appData, "TheCopyfather", "llama-server.exe"),
		)
	}

	// 2. Executable-relative engine dir
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, "engine", "llama-server.exe"),
			filepath.Join(exeDir, "llama-server.exe"),
		)
	}

	// 3. Working directory engine dir
	cwd, err := os.Getwd()
	if err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "engine", "llama-server.exe"),
			filepath.Join(cwd, "bin", "llama-server.exe"),
			filepath.Join(cwd, "llama-server.exe"),
		)
	}

	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, true
		}
	}
	return GetDefaultBinaryPath(), false
}

// FindModel searches for the GGUF model file with AppData as the primary standard location
func FindModel(customPath string) (string, bool) {
	candidates := []string{}
	if customPath != "" {
		candidates = append(candidates, customPath)
	}

	// Common filenames for SmolLM2 & Qwen GGUF models
	modelFiles := []string{
		"smollm2-360m-instruct-q4_k_m.gguf",
		"SmolLM2-360M-Instruct-Q4_K_M.gguf",
		"smollm2-360m.gguf",
		"SmolLM2-360M.gguf",
		"qwen3-1.7b-q4_k_m.gguf",
		"qwen3-1.7b.Q4_K_M.gguf",
		"qwen3-1.7b-instruct-q4_k_m.gguf",
		"qwen3-1.7b-instruct.Q4_K_M.gguf",
		"qwen3-1.7b.gguf",
		"qwen2.5-1.5b-instruct-q4_k_m.gguf",
		"qwen2.5-1.5b.Q4_K_M.gguf",
		"qwen2.5-1.5b-instruct.gguf",
		"qwen2.5-1.5b.gguf",
		"qwen2.5-0.5b-instruct-q4_k_m.gguf",
		"qwen2.5-0.5b.Q4_K_M.gguf",
		"model.gguf",
	}

	// 1. AppData directory models/ (PRIMARY for installed/distributed app)
	appData := os.Getenv("APPDATA")
	if appData != "" {
		for _, name := range modelFiles {
			candidates = append(candidates, filepath.Join(appData, "TheCopyfather", "models", name))
		}
	}

	// 2. Executable-relative models/
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		for _, name := range modelFiles {
			candidates = append(candidates, filepath.Join(exeDir, "models", name))
		}
	}

	// 3. Working directory models/
	cwd, err := os.Getwd()
	if err == nil {
		for _, name := range modelFiles {
			candidates = append(candidates, filepath.Join(cwd, "models", name))
		}
	}

	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, true
		}
	}

	// Fallback scan: Search for ANY .gguf file in the model directories
	searchDirs := []string{}
	if appData != "" {
		searchDirs = append(searchDirs, filepath.Join(appData, "TheCopyfather", "models"))
	}
	if exePath != "" {
		searchDirs = append(searchDirs, filepath.Join(filepath.Dir(exePath), "models"))
	}
	if cwd != "" {
		searchDirs = append(searchDirs, filepath.Join(cwd, "models"), filepath.Join(cwd, "engine"))
	}

	for _, dir := range searchDirs {
		entries, err := os.ReadDir(dir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".gguf") {
					return filepath.Join(dir, entry.Name()), true
				}
			}
		}
	}

	return GetDefaultModelPath(), false
}

// GetStatus returns the current engine status
func (e *Engine) GetStatus() Status {
	e.mu.Lock()
	defer e.mu.Unlock()

	binPath, binFound := FindBinary(e.cfg.BinaryPath)
	modelPath, modelFound := FindModel(e.cfg.ModelPath)
	threads := DetectThreads(e.cfg.CPUThreads, e.cfg.MaxCPUThreads)

	status := Status{
		IsRunning:     e.isRunning && isPortOpen(e.actualPort),
		ModelLoaded:   filepath.Base(modelPath),
		ModelFound:    modelFound,
		ModelPath:     modelPath,
		BinaryFound:   binFound,
		BinaryPath:    binPath,
		Port:          e.actualPort,
		Threads:       threads,
		MaxThreads:    e.cfg.MaxCPUThreads,
		ContextSize:   e.cfg.ContextSize,
		Hardware:      "CPU",
		ThinkingOff:   e.cfg.DisableThinking,
		StreamingOn:   e.cfg.Streaming,
	}

	if !binFound {
		status.Error = fmt.Sprintf("llama-server.exe binary not found at %s", binPath)
	} else if !modelFound {
		status.Error = fmt.Sprintf("GGUF model not found at %s", modelPath)
	}

	return status
}

// Start launches the embedded llama-server process
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.isRunning && isPortOpen(e.actualPort) {
		return nil // already running
	}

	binPath, binFound := FindBinary(e.cfg.BinaryPath)
	if !binFound {
		return fmt.Errorf("llama-server.exe not found at %s. Please place llama-server.exe in the engine folder", binPath)
	}

	modelPath, modelFound := FindModel(e.cfg.ModelPath)
	if !modelFound {
		return fmt.Errorf("model file not found at %s. Please place Qwen3-1.7B-Q4_K_M.gguf in the models folder", modelPath)
	}

	// Determine port
	port := e.cfg.Port
	if port <= 0 {
		port = 8085
	}
	e.actualPort = port

	// Calculate CPU threads
	threads := DetectThreads(e.cfg.CPUThreads, e.cfg.MaxCPUThreads)
	e.threads = threads

	// Context size: limit to 2048 - 4096
	ctxSize := e.cfg.ContextSize
	if ctxSize < 2048 {
		ctxSize = 2048
	} else if ctxSize > 4096 {
		ctxSize = 4096
	}

	// Arguments for llama.cpp llama-server
	// -m: Model path
	// -c: Context size (2K-4K)
	// -t: Threads
	// -np 1: Single slot to minimize memory footprint
	// -ngl 0: Hardware CPU (0 layers offloaded to GPU)
	// --host: 127.0.0.1
	// --port: port
	args := []string{
		"-m", modelPath,
		"-c", fmt.Sprintf("%d", ctxSize),
		"-t", fmt.Sprintf("%d", threads),
		"-np", "1", // Single slot for personal desktop app to minimize RAM
		"-ngl", "0", // Hardware: CPU
		"--host", "127.0.0.1",
		"--port", fmt.Sprintf("%d", port),
	}

	if e.cfg.DisableThinking {
		args = append(args, "--reasoning-budget", "0")
	}

	cmd := exec.Command(binPath, args...)

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	// On Windows, hide command window completely so it runs silently in the background
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start embedded llama-server: %w", err)
	}

	e.cmd = cmd
	e.isRunning = true

	processDone := make(chan error, 1)
	go func() {
		err := cmd.Wait()
		e.mu.Lock()
		e.isRunning = false
		if err != nil {
			e.lastError = strings.TrimSpace(stderrBuf.String())
		}
		e.mu.Unlock()
		processDone <- err
	}()

	return e.waitForReady(port, processDone, &stderrBuf, 15*time.Second)
}

// waitForReady polls the local endpoint until it responds or process exits
func (e *Engine) waitForReady(port int, processDone <-chan error, stderrBuf *bytes.Buffer, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 800 * time.Millisecond}
	url := fmt.Sprintf("http://127.0.0.1:%d/health", port)

	for time.Now().Before(deadline) {
		select {
		case err := <-processDone:
			errMsg := strings.TrimSpace(stderrBuf.String())
			if errMsg != "" {
				return fmt.Errorf("llama-server exited unexpectedly: %s", errMsg)
			}
			return fmt.Errorf("llama-server exited with code: %v", err)
		default:
		}

		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusServiceUnavailable {
				// Server is up and responding
				return nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}

	select {
	case err := <-processDone:
		errMsg := strings.TrimSpace(stderrBuf.String())
		if errMsg != "" {
			return fmt.Errorf("llama-server exited: %s", errMsg)
		}
		return fmt.Errorf("llama-server exited with code: %v", err)
	default:
	}

	return fmt.Errorf("llama-server started but did not respond on port %d within %v", port, timeout)
}

// Stop cleanly terminates the embedded llama-server process
func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.cmd != nil && e.cmd.Process != nil {
		_ = e.cmd.Process.Kill()
		e.cmd = nil
	}
	e.isRunning = false
	return nil
}

// GetBaseURL returns the local OpenAI-compatible endpoint URL for this engine
func (e *Engine) GetBaseURL() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	port := e.actualPort
	if port <= 0 {
		port = e.cfg.Port
		if port <= 0 {
			port = 8085
		}
	}
	return fmt.Sprintf("http://127.0.0.1:%d/v1", port)
}

// IsRunning returns whether the engine is active
func (e *Engine) IsRunning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.isRunning && isPortOpen(e.actualPort)
}

// isPortOpen checks if a TCP port is currently open
func isPortOpen(port int) bool {
	if port <= 0 {
		return false
	}
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 200*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// StripThinking removes any reasoning / <think>...</think> tags from output text
func StripThinking(text string) string {
	// Remove <think>...</think> blocks
	for {
		start := indexCaseInsensitive(text, "<think>")
		if start == -1 {
			break
		}
		end := indexCaseInsensitive(text, "</think>")
		if end != -1 {
			text = text[:start] + text[end+8:]
		} else {
			// Unclosed <think> tag, strip everything from <think>
			text = text[:start]
			break
		}
	}
	return text
}

func indexCaseInsensitive(s, substr string) int {
	sLower := []rune(filepath.ToSlash(s)) // fast lower roughly
	_ = sLower
	idx := -1
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFold(s[i:i+len(substr)], substr) {
			return i
		}
	}
	return idx
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
