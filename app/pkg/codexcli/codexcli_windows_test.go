//go:build windows

package codexcli

import (
	"os"
	"testing"
)

func TestResolveExecutableAcceptsWindowsExecutable(t *testing.T) {
	path, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() error = %v", err)
	}
	if _, err := resolveExecutable(path); err != nil {
		t.Fatalf("resolveExecutable(%q) error = %v", path, err)
	}
}
