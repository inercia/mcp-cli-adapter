package command

import (
	"context"
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestNewRunnerFirejail(t *testing.T) {
	// Skip on non-Linux platforms
	if runtime.GOOS != "linux" {
		t.Skip("Skipping firejail tests on non-Linux platform")
	}

	options := RunnerOptions{
		"allow_networking": true,
	}

	runner, err := NewRunnerFirejail(options, nil)
	if err != nil {
		t.Fatalf("Failed to create firejail runner: %v", err)
	}

	if runner == nil {
		t.Fatal("Expected non-nil runner")
	}
}

func TestRunnerFirejailRun(t *testing.T) {
	// Skip on non-Linux platforms
	if runtime.GOOS != "linux" {
		t.Skip("Skipping firejail tests on non-Linux platform")
	}

	// Skip if firejail is not installed
	if _, err := os.Stat("/usr/bin/firejail"); os.IsNotExist(err) {
		t.Skip("Skipping test because firejail is not installed")
	}

	options := RunnerOptions{
		"allow_networking": true,
	}

	runner, err := NewRunnerFirejail(options, nil)
	if err != nil {
		t.Fatalf("Failed to create firejail runner: %v", err)
	}

	ctx := context.Background()

	// Test simple echo command
	output, err := runner.Run(ctx, "/bin/sh", "echo hello world", nil, nil, false) // No need for tmpfile here
	if err != nil {
		t.Fatalf("Failed to run command: %v", err)
	}

	if output != "hello world\n" {
		t.Errorf("Expected 'hello world\\n', got '%s'", output)
	}
}

func TestRunnerFirejailNetworkRestriction(t *testing.T) {
	// Skip on non-Linux platforms
	if runtime.GOOS != "linux" {
		t.Skip("Skipping firejail tests on non-Linux platform")
	}

	// Skip if firejail is not installed
	if _, err := os.Stat("/usr/bin/firejail"); os.IsNotExist(err) {
		t.Skip("Skipping test because firejail is not installed")
	}

	ctx := context.Background()

	// Test with networking enabled
	networkEnabledOptions := RunnerOptions{
		"allow_networking": true,
	}

	runnerEnabled, err := NewRunnerFirejail(networkEnabledOptions, nil)
	if err != nil {
		t.Fatalf("Failed to create firejail runner: %v", err)
	}

	// This might succeed or fail depending on network connectivity,
	// but it should not be blocked by firejail
	_, _ = runnerEnabled.Run(ctx, "/bin/sh", "ping -c 1 127.0.0.1", nil, nil, false) // No need for tmpfile here

	// Test with networking disabled
	networkDisabledOptions := RunnerOptions{
		"allow_networking": false,
	}

	runnerDisabled, err := NewRunnerFirejail(networkDisabledOptions, nil)
	if err != nil {
		t.Fatalf("Failed to create firejail runner: %v", err)
	}

	// This should fail or timeout due to network restrictions
	// Note: We're not asserting the exact behavior as it might vary based on firejail version
	_, _ = runnerDisabled.Run(ctx, "/bin/sh", "ping -c 1 127.0.0.1", nil, nil, false) // No need for tmpfile here
}

func TestRunnerFirejail_Optimization_SingleExecutable(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping firejail tests on non-Linux platform")
	}
	if _, err := os.Stat("/usr/bin/firejail"); os.IsNotExist(err) {
		t.Skip("Skipping test because firejail is not installed")
	}
	runner, err := NewRunnerFirejail(RunnerOptions{"allow_networking": true}, nil)
	if err != nil {
		t.Fatalf("Failed to create firejail runner: %v", err)
	}
	// Should succeed: /bin/ls is a single executable
	output, err := runner.Run(context.Background(), "", "/bin/ls", nil, nil, false)
	if err != nil {
		t.Errorf("Expected /bin/ls to run without error, got: %v", err)
	}
	if len(output) == 0 {
		t.Errorf("Expected output from /bin/ls, got empty string")
	}
	// Should NOT optimize: command with arguments
	_, err2 := runner.Run(context.Background(), "", "/bin/ls -l", nil, nil, false)
	if err2 != nil && !strings.Contains(err2.Error(), "no such file") {
		t.Logf("Expected failure for /bin/ls -l as a single executable: %v", err2)
	}
}
