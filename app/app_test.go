package main

import (
	"bytes"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGxReadLogReturnsMostRecentLines verifies that the requested trailing entries are returned.
func TestGxReadLogReturnsMostRecentLines(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "helaix.log")
	contents := "line 1\nline 2\nline 3\nline 4"
	if err := os.WriteFile(logPath, []byte(contents), 0600); err != nil {
		t.Fatalf("write test log: %v", err)
	}

	app := &App{logPath: logPath}
	got, err := app.GxReadLog(2)
	if err != nil {
		t.Fatalf("GxReadLog() error = %v", err)
	}
	if want := "line 3\nline 4"; got != want {
		t.Errorf("GxReadLog() = %q, want %q", got, want)
	}
}

// TestGxReadLogReadsTailOfOversizedFile verifies bounded tail reading for large historical logs.
func TestGxReadLogReadsTailOfOversizedFile(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "helaix.log")
	contents := strings.Repeat("old entry\n", int(maxLogBytes/10)) + "latest 1\nlatest 2\nlatest 3"
	if err := os.WriteFile(logPath, []byte(contents), 0600); err != nil {
		t.Fatalf("write test log: %v", err)
	}

	app := &App{logPath: logPath}
	got, err := app.GxReadLog(2)
	if err != nil {
		t.Fatalf("GxReadLog() error = %v", err)
	}
	if want := "latest 2\nlatest 3"; got != want {
		t.Errorf("GxReadLog() = %q, want %q", got, want)
	}
}

// TestGxReadLogReportsInitializationFailure verifies setup errors reach the diagnostics caller.
func TestGxReadLogReportsInitializationFailure(t *testing.T) {
	app := &App{logPath: "log path", logInitErr: errors.New("permission denied")}
	if path, err := app.GxGetLogPath(); path != "log path" || err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("GxGetLogPath() = (%q, %v), want path and initialization error", path, err)
	}
	if _, err := app.GxReadLog(200); err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("GxReadLog() error = %v, want initialization error", err)
	}
	if err := app.GxClearLog(); err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("GxClearLog() error = %v, want initialization error", err)
	}
}

// TestLogAIErrorBoundsFileGrowth verifies rotation and per-entry size limits on append.
func TestLogAIErrorBoundsFileGrowth(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "helaix.log")
	if err := os.WriteFile(logPath, []byte(strings.Repeat("x", int(maxLogBytes))), 0600); err != nil {
		t.Fatalf("write test log: %v", err)
	}

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatalf("open test log: %v", err)
	}
	defer file.Close()

	app := &App{
		logger:  log.New(file, "helaix ", log.LstdFlags),
		logFile: file,
		logPath: logPath,
	}
	app.logAIError("test", "model", errors.New(strings.Repeat("x", int(maxLogBytes)+1)))

	info, err := file.Stat()
	if err != nil {
		t.Fatalf("stat test log: %v", err)
	}
	if info.Size() > maxLogBytes {
		t.Errorf("log size = %d, want at most %d", info.Size(), maxLogBytes)
	}
	if info.Size() > maxLogEntryBytes+64 {
		t.Errorf("single log entry size = %d, want at most %d", info.Size(), maxLogEntryBytes+64)
	}
}

// TestLogAIErrorOmitsRawResponses verifies model output is not copied into application logs.
func TestLogAIErrorOmitsRawResponses(t *testing.T) {
	var output bytes.Buffer
	app := &App{logger: log.New(&output, "", 0)}
	app.logAIError("preset_engineer", "model", errors.New("parse failure. Raw: private generated content"))

	if strings.Contains(output.String(), "private generated content") {
		t.Fatal("log entry contains raw model output")
	}
	if !strings.Contains(output.String(), "raw response omitted") {
		t.Errorf("log entry = %q, want redaction marker", output.String())
	}
}

// TestGxClearLogTruncatesFile verifies clearing retains an empty log file.
func TestGxClearLogTruncatesFile(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "helaix.log")
	if err := os.WriteFile(logPath, []byte("log entry"), 0600); err != nil {
		t.Fatalf("write test log: %v", err)
	}

	app := &App{logPath: logPath}
	if err := app.GxClearLog(); err != nil {
		t.Fatalf("GxClearLog() error = %v", err)
	}

	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("stat test log: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("log size = %d, want 0", info.Size())
	}
}
