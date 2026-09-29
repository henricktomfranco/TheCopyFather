package engine

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SetupStatus represents the state of required engine and model files
type SetupStatus struct {
	EngineInstalled bool    `json:"engine_installed"`
	EnginePath      string  `json:"engine_path"`
	ModelInstalled  bool    `json:"model_installed"`
	ModelPath       string  `json:"model_path"`
	IsDownloading   bool    `json:"is_downloading"`
	CurrentStep     string  `json:"current_step"` // "idle", "engine", "extracting", "model", "completed", "error"
	ProgressPercent float64 `json:"progress_percent"`
	BytesDownloaded int64   `json:"bytes_downloaded"`
	TotalBytes      int64   `json:"total_bytes"`
	Speed           string  `json:"speed"`
	Error           string  `json:"error,omitempty"`
}

// Downloader manages downloading missing engine and model files
type Downloader struct {
	mu         sync.Mutex
	status     SetupStatus
	cancelFunc context.CancelFunc
}

var (
	defaultDownloader *Downloader
	downloaderOnce    sync.Once
)

// GetDownloader returns the global Downloader instance
func GetDownloader() *Downloader {
	downloaderOnce.Do(func() {
		defaultDownloader = &Downloader{
			status: SetupStatus{
				CurrentStep: "idle",
			},
		}
	})
	return defaultDownloader
}

// CheckSetupStatus inspects if engine and model files exist
func (d *Downloader) CheckSetupStatus() SetupStatus {
	d.mu.Lock()
	defer d.mu.Unlock()

	binPath, binFound := FindBinary("")
	modelPath, modelFound := FindModel("")

	d.status.EnginePath = binPath
	d.status.EngineInstalled = binFound
	d.status.ModelPath = modelPath
	d.status.ModelInstalled = modelFound

	return d.status
}

// ProgressFunc callback for reporting download progress
type ProgressFunc func(status SetupStatus)

// DownloadAll missing files automatically
func (d *Downloader) DownloadAll(ctx context.Context, onProgress ProgressFunc) error {
	d.mu.Lock()
	if d.status.IsDownloading {
		d.mu.Unlock()
		return fmt.Errorf("a download is already in progress")
	}

	EnsureDirectories()

	ctx, cancel := context.WithCancel(ctx)
	d.cancelFunc = cancel
	d.status.IsDownloading = true
	d.status.Error = ""
	d.mu.Unlock()

	defer func() {
		d.mu.Lock()
		d.status.IsDownloading = false
		d.cancelFunc = nil
		d.mu.Unlock()
	}()

	updateStatus := func(step string, downloaded, total int64, percent float64, speed string, errStr string) {
		d.mu.Lock()
		d.status.CurrentStep = step
		d.status.BytesDownloaded = downloaded
		d.status.TotalBytes = total
		d.status.ProgressPercent = percent
		d.status.Speed = speed
		d.status.Error = errStr
		current := d.status
		d.mu.Unlock()

		if onProgress != nil {
			onProgress(current)
		}
	}

	// 1. Check and download Engine if missing
	binPath, binFound := FindBinary("")
	if !binFound {
		updateStatus("engine", 0, 0, 0, "", "")
		if err := d.downloadEngine(ctx, updateStatus); err != nil {
			updateStatus("error", 0, 0, 0, "", err.Error())
			return err
		}
		binPath, _ = FindBinary("")
	}
	d.mu.Lock()
	d.status.EngineInstalled = true
	d.status.EnginePath = binPath
	d.mu.Unlock()

	// 2. Check and download Model if missing
	modelPath, modelFound := FindModel("")
	if !modelFound {
		updateStatus("model", 0, 0, 0, "", "")
		if err := d.downloadModel(ctx, updateStatus); err != nil {
			updateStatus("error", 0, 0, 0, "", err.Error())
			return err
		}
		modelPath, _ = FindModel("")
	}
	d.mu.Lock()
	d.status.ModelInstalled = true
	d.status.ModelPath = modelPath
	d.mu.Unlock()

	updateStatus("completed", 0, 0, 100, "", "")
	return nil
}

// Cancel terminates any running download
func (d *Downloader) Cancel() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cancelFunc != nil {
		d.cancelFunc()
	}
}

// downloadEngine fetches and extracts llama.cpp release
func (d *Downloader) downloadEngine(ctx context.Context, updateStatus func(string, int64, int64, float64, string, string)) error {
	engineDir := GetDefaultEngineDir()
	_ = os.MkdirAll(engineDir, 0755)

	downloadURL, fileName, err := getLatestLlamaCppReleaseURL()
	if err != nil {
		// Fallback to verified reliable release asset
		downloadURL = "https://github.com/ggml-org/llama.cpp/releases/download/b11227/llama-b11227-bin-win-cpu-x64.zip"
		fileName = "llama-b11227-bin-win-cpu-x64.zip"
	}

	tempZipPath := filepath.Join(os.TempDir(), fileName)
	defer os.Remove(tempZipPath)

	err = downloadFileWithProgress(ctx, downloadURL, tempZipPath, func(dl, total int64, pct float64, spd string) {
		updateStatus("engine", dl, total, pct, spd, "")
	})
	if err != nil {
		return fmt.Errorf("failed downloading engine: %w", err)
	}

	updateStatus("extracting", 0, 0, 100, "", "")
	if err := unzip(tempZipPath, engineDir); err != nil {
		return fmt.Errorf("failed extracting engine zip: %w", err)
	}

	return nil
}

// downloadModel fetches Qwen3 GGUF model
func (d *Downloader) downloadModel(ctx context.Context, updateStatus func(string, int64, int64, float64, string, string)) error {
	modelsDir := GetDefaultModelsDir()
	_ = os.MkdirAll(modelsDir, 0755)

	targetFile := filepath.Join(modelsDir, "smollm2-360m-instruct-q4_k_m.gguf")
	tempFile := targetFile + ".tmp"
	defer os.Remove(tempFile)

	// Primary verified Hugging Face download URL for SmolLM2-360M-Instruct Q4_K_M GGUF (258 MB)
	modelURL := "https://huggingface.co/bartowski/SmolLM2-360M-Instruct-GGUF/resolve/main/SmolLM2-360M-Instruct-Q4_K_M.gguf"

	err := downloadFileWithProgress(ctx, modelURL, tempFile, func(dl, total int64, pct float64, spd string) {
		updateStatus("model", dl, total, pct, spd, "")
	})
	if err != nil {
		return fmt.Errorf("failed downloading model: %w", err)
	}

	// Rename temp file to final destination (remove existing target first on Windows)
	_ = os.Remove(targetFile)
	if err := os.Rename(tempFile, targetFile); err != nil {
		return fmt.Errorf("failed renaming downloaded model: %w", err)
	}

	return nil
}

// downloadFileWithProgress performs HTTP GET with progress tracking
func downloadFileWithProgress(ctx context.Context, url, destPath string, onProgress func(dl, total int64, pct float64, spd string)) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "TheCopyFather")

	client := &http.Client{
		Timeout: 0, // No global timeout for large file downloads
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download server returned HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	totalBytes := resp.ContentLength
	var downloaded int64

	buf := make([]byte, 64*1024)
	lastReport := time.Now()
	lastBytes := int64(0)
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, rErr := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := out.Write(buf[:n]); wErr != nil {
				return wErr
			}
			downloaded += int64(n)

			now := time.Now()
			if now.Sub(lastReport) >= 200*time.Millisecond || rErr != nil {
				pct := float64(0)
				if totalBytes > 0 {
					pct = float64(downloaded) / float64(totalBytes) * 100.0
				}
				elapsedSec := now.Sub(lastReport).Seconds()
				speedBytes := float64(downloaded-lastBytes) / elapsedSec
				if elapsedSec == 0 {
					speedBytes = float64(downloaded) / now.Sub(startTime).Seconds()
				}
				speedStr := formatSpeed(speedBytes)

				if onProgress != nil {
					onProgress(downloaded, totalBytes, pct, speedStr)
				}
				lastReport = now
				lastBytes = downloaded
			}
		}

		if rErr == io.EOF {
			break
		}
		if rErr != nil {
			return rErr
		}
	}

	return nil
}

// formatSpeed formats bytes per second into human-readable string
func formatSpeed(bytesPerSec float64) string {
	if bytesPerSec < 1024 {
		return fmt.Sprintf("%.0f B/s", bytesPerSec)
	} else if bytesPerSec < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", bytesPerSec/1024)
	}
	return fmt.Sprintf("%.1f MB/s", bytesPerSec/(1024*1024))
}

// unzip unpacks a zip archive into destDir
func unzip(srcZip, destDir string) error {
	r, err := zip.OpenReader(srcZip)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		targetPath := filepath.Join(destDir, f.Name)

		// Check for ZipSlip
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(destDir)) {
			continue
		}

		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(targetPath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// getLatestLlamaCppReleaseURL queries GitHub for latest CPU release
func getLatestLlamaCppReleaseURL() (string, string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/repos/ggml-org/llama.cpp/releases?per_page=1", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "TheCopyFather")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releases []struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", "", err
	}

	if len(releases) == 0 {
		return "", "", fmt.Errorf("no releases found")
	}

	for _, asset := range releases[0].Assets {
		if strings.HasSuffix(asset.Name, "bin-win-cpu-x64.zip") {
			return asset.BrowserDownloadURL, asset.Name, nil
		}
	}

	return "", "", fmt.Errorf("cpu release asset not found")
}
