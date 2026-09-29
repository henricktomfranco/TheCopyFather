package engine

import (
	"context"
	"os"
	"testing"
	"textrewriter/internal/ollama"
	"time"
)

func TestDetectThreads(t *testing.T) {
	threads := DetectThreads(8, 0)
	if threads != 8 {
		t.Errorf("Expected 8 threads, got %d", threads)
	}

	threads = DetectThreads(0, 4)
	if threads > 4 {
		t.Errorf("Expected threads <= 4, got %d", threads)
	}
	if threads < 1 {
		t.Errorf("Expected threads >= 1, got %d", threads)
	}
}

func TestFindBinaryAndModel(t *testing.T) {
	binPath, binFound := FindBinary("")
	if binPath == "" {
		t.Error("Expected non-empty binary path")
	}

	modelPath, modelFound := FindModel("")
	if modelPath == "" {
		t.Error("Expected non-empty model path")
	}

	t.Logf("FindBinary: %s (found: %v)", binPath, binFound)
	t.Logf("FindModel: %s (found: %v)", modelPath, modelFound)
}

func TestEnsureDirectories(t *testing.T) {
	EnsureDirectories()

	engineDir := GetDefaultEngineDir()
	modelsDir := GetDefaultModelsDir()

	if _, err := os.Stat(engineDir); os.IsNotExist(err) {
		t.Errorf("Engine dir was not created: %s", engineDir)
	}
	if _, err := os.Stat(modelsDir); os.IsNotExist(err) {
		t.Errorf("Models dir was not created: %s", modelsDir)
	}
}

func TestCheckSetupStatus(t *testing.T) {
	downloader := GetDownloader()
	status := downloader.CheckSetupStatus()

	if status.EnginePath == "" {
		t.Error("Expected non-empty EnginePath")
	}
	if status.ModelPath == "" {
		t.Error("Expected non-empty ModelPath")
	}

	t.Logf("SetupStatus: EngineInstalled=%v, ModelInstalled=%v", status.EngineInstalled, status.ModelInstalled)
}

func TestEngineLifecycle(t *testing.T) {
	eng := New(Config{
		ContextSize:     2048,
		CPUThreads:      2,
		DisableThinking: true,
		Streaming:       true,
	})

	status := eng.GetStatus()
	if !status.BinaryFound || !status.ModelFound {
		t.Skip("Binary or model not present, skipping live engine test")
	}

	if err := eng.Start(); err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}
	defer eng.Stop()

	if !eng.IsRunning() {
		t.Fatal("Expected engine to be running")
	}

	t.Log("Embedded llama-server successfully started and verified healthy!")
}

func TestLiveRewrite(t *testing.T) {
	eng := New(Config{
		ContextSize:     2048,
		CPUThreads:      2,
		DisableThinking: true,
		Streaming:       true,
	})

	if err := eng.Start(); err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}
	defer eng.Stop()

	client := ollama.NewOpenAICompatibleClientWithAllOptions(
		eng.GetBaseURL(),
		"Qwen3-1.7B",
		"",
		false,
		true,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	result, err := client.GenerateRewrite(ctx, "hello world this is a test", "grammar", "Fix any typos.")
	if err != nil {
		t.Fatalf("GenerateRewrite failed: %v", err)
	}

	t.Logf("Live rewrite completed successfully! Output: %s", result)
}
