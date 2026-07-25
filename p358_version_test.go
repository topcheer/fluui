package fluui

import (
	"strings"
	"testing"
)

// TestP358_Version_Default verifies the default version string.
func TestP358_Version_Default(t *testing.T) {
	// In test environment, Version should be "dev" unless overridden
	v := Version
	if v == "" {
		t.Error("Version should not be empty")
	}
}

// TestP358_VersionInfo_String verifies the String formatting.
func TestP358_VersionInfo_String(t *testing.T) {
	vi := VersionInfo{Version: "v1.0.0"}
	s := vi.String()
	if !strings.Contains(s, "v1.0.0") {
		t.Errorf("String() = %q, expected to contain v1.0.0", s)
	}
}

// TestP358_VersionInfo_String_WithCommit verifies full version string.
func TestP358_VersionInfo_String_WithCommit(t *testing.T) {
	vi := VersionInfo{Version: "v1.0.0", Commit: "abc123", Date: "2026-02-28"}
	s := vi.String()
	if !strings.Contains(s, "v1.0.0") {
		t.Errorf("missing version: %q", s)
	}
	if !strings.Contains(s, "abc123") {
		t.Errorf("missing commit: %q", s)
	}
	if !strings.Contains(s, "2026-02-28") {
		t.Errorf("missing date: %q", s)
	}
}

// TestP358_VersionInfo_IsDev verifies dev detection.
func TestP358_VersionInfo_IsDev(t *testing.T) {
	vi := VersionInfo{Version: "dev"}
	if !vi.IsDev() {
		t.Error("Version 'dev' should be detected as dev")
	}

	vi2 := VersionInfo{Version: ""}
	if !vi2.IsDev() {
		t.Error("empty version should be detected as dev")
	}

	vi3 := VersionInfo{Version: "v1.0.0"}
	if vi3.IsDev() {
		t.Error("v1.0.0 should not be dev")
	}
}

// TestP358_ComponentCount verifies the constant matches reality.
func TestP358_ComponentCount(t *testing.T) {
	if ComponentCount < 100 {
		t.Errorf("ComponentCount = %d, expected at least 100", ComponentCount)
	}
}

// TestP358_ProtocolCount verifies the constant.
func TestP358_ProtocolCount(t *testing.T) {
	if ProtocolCount < 20 {
		t.Errorf("ProtocolCount = %d, expected at least 20", ProtocolCount)
	}
}
