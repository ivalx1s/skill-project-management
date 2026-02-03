package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestShowCommand(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	showCmd.SetOut(&buf)

	// Capture stdout
	old := showCmd.OutOrStdout()
	_ = old

	err := runShow(showCmd, []string{testTask1ID})
	if err != nil {
		t.Fatalf("runShow: %v", err)
	}
}

func TestShowNotFound(t *testing.T) {
	bd := setupTestBoard(t)
	boardDir = bd

	err := runShow(showCmd, []string{"TASK-999"})
	if err == nil {
		t.Fatal("expected error for missing element")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want 'not found'", err.Error())
	}
}
