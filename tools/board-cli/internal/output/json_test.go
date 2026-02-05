package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "value"}

	err := PrintJSON(&buf, data)
	if err != nil {
		t.Fatalf("PrintJSON error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("expected key=value, got %q", result["key"])
	}
}

func TestPrintError(t *testing.T) {
	var buf bytes.Buffer

	err := PrintError(&buf, NotFound, "Element TASK-999 not found", nil)
	if err != nil {
		t.Fatalf("PrintError error: %v", err)
	}

	var result JSONError
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if result.Error.Code != NotFound {
		t.Errorf("expected code NOT_FOUND, got %q", result.Error.Code)
	}
	if result.Error.Message != "Element TASK-999 not found" {
		t.Errorf("unexpected message: %q", result.Error.Message)
	}
	if result.Error.Details == nil {
		t.Error("Details should be non-nil empty map")
	}
}

func TestPrintErrorWithDetails(t *testing.T) {
	var buf bytes.Buffer

	details := map[string]interface{}{
		"elementId": "TASK-999",
		"status":    "blocked",
	}
	err := PrintError(&buf, InvalidStatus, "Invalid status transition", details)
	if err != nil {
		t.Fatalf("PrintError error: %v", err)
	}

	var result JSONError
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if result.Error.Code != InvalidStatus {
		t.Errorf("expected code INVALID_STATUS, got %q", result.Error.Code)
	}
	if result.Error.Details["elementId"] != "TASK-999" {
		t.Errorf("expected elementId=TASK-999, got %v", result.Error.Details["elementId"])
	}
}

func TestNewErrorResponse(t *testing.T) {
	resp := NewErrorResponse(CycleDetected, "Dependency cycle detected", nil)

	if resp.Error.Code != CycleDetected {
		t.Errorf("expected code CYCLE_DETECTED, got %q", resp.Error.Code)
	}
	if resp.Error.Details == nil {
		t.Error("Details should be non-nil empty map")
	}
}

func TestMarshalJSON(t *testing.T) {
	data := map[string]int{"count": 42}
	bytes, err := MarshalJSON(data)
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	// Should be indented
	if !contains(string(bytes), "\n") {
		t.Error("expected indented JSON output")
	}
}

func TestErrorCodes(t *testing.T) {
	// Verify all error codes are defined correctly
	codes := []ErrorCode{
		NotFound,
		InvalidID,
		InvalidStatus,
		CycleDetected,
		ValidationError,
		InternalError,
	}

	expected := []string{
		"NOT_FOUND",
		"INVALID_ID",
		"INVALID_STATUS",
		"CYCLE_DETECTED",
		"VALIDATION_ERROR",
		"INTERNAL_ERROR",
	}

	for i, code := range codes {
		if string(code) != expected[i] {
			t.Errorf("ErrorCode %d: expected %q, got %q", i, expected[i], code)
		}
	}
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
