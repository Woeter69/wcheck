package cmd

import (
	"os/exec"
	"strings"
	"testing"
)

func TestHelpCommand(t *testing.T) {
	cmd := exec.Command("go", "run", "../main.go", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Help command failed: %v\nOutput: %s", err, string(output))
	}

	expected := "wcheck is a headless browser scanner"
	if !strings.Contains(string(output), expected) {
		t.Errorf("Expected output to contain '%s', got: %s", expected, string(output))
	}
}

func TestInvalidURL(t *testing.T) {
	// Running against a non-existent local port should fail with exit status 1
	cmd := exec.Command("go", "run", "../main.go", "scan", "http://localhost:9999")
	err := cmd.Run()
	
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}

	if exitError, ok := err.(*exec.ExitError); ok {
		if exitError.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got %d", exitError.ExitCode())
		}
	} else {
		t.Errorf("Expected ExitError, got %T", err)
	}
}
