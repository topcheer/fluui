package mock

import (
	"strings"
	"testing"
)

func TestCollectAllChunks_ParseError_P251(t *testing.T) {
	// Create reader with invalid SSE data
	r := NewSSEStreamReader(strings.NewReader("data: not-json\n\n"))
	_, err := CollectAllChunks(r)
	if err == nil {
		t.Error("expected parse error for invalid JSON")
	}
}

func TestCollectAllContent_ParseError_P251(t *testing.T) {
	r := NewSSEStreamReader(strings.NewReader("data: not-json\n\n"))
	_, err := CollectAllContent(r)
	if err == nil {
		t.Error("expected parse error")
	}
}

func TestCollectAllChunks_ReadError_P251(t *testing.T) {
	// Reader that always errors
	r := NewSSEStreamReader(strings.NewReader(""))
	// EOF is expected, not error
	chunks, err := CollectAllChunks(r)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks, got %d", len(chunks))
	}
}
