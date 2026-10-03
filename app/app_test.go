package main

import (
	"os"
	"path/filepath"
	"testing"
)

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
