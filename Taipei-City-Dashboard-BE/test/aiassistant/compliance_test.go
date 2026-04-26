package aiassistant_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestTWCCModelAndRateLimitCompliance(t *testing.T) {
	beRoot := backendRoot(t)
	assertFileContains(t, filepath.Join(beRoot, "global", "global.go"),
		`Model:         getEnv("TWCC_MODEL", "llama3.3-ffm-70b-16k-chat")`)
	assertFileNotContains(t, filepath.Join(beRoot, "global", "global.go"),
		"llama3.3-ffm-70b-32k-chat")
	assertFileContains(t, filepath.Join(beRoot, "global", "consts.go"),
		"AIChatLimitAPIRequestsTimes        = 30")
	assertFileContains(t, filepath.Join(beRoot, "app", "routes", "aiRoutes.go"),
		"LimitAPIRequests(global.AIChatLimitAPIRequestsTimes")
}

func TestFrontendDoesNotContainTWCCProviderAccess(t *testing.T) {
	feSrc := filepath.Join(filepath.Dir(backendRoot(t)), "Taipei-City-Dashboard-FE", "src")
	banned := []string{
		"api-ams.twcc.ai",
		"TWCC_API_KEY",
		"TWCC_MODEL",
		"llama3.3",
		"api.openai",
		"api.anthropic",
	}
	err := filepath.WalkDir(feSrc, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		for _, token := range banned {
			if strings.Contains(content, token) {
				t.Fatalf("frontend file %s contains banned token %q", path, token)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk frontend src: %v", err)
	}
}

func backendRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func assertFileContains(t *testing.T, path string, token string) {
	t.Helper()
	if !strings.Contains(readFile(t, path), token) {
		t.Fatalf("%s missing %q", path, token)
	}
}

func assertFileNotContains(t *testing.T, path string, token string) {
	t.Helper()
	if strings.Contains(readFile(t, path), token) {
		t.Fatalf("%s contains %q", path, token)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
