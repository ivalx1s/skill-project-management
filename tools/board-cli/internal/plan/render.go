package plan

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aagrigore/task-board/internal/board"
)

// RenderDOT writes a DOT string to a temporary file, invokes the `dot` binary
// to render it, and writes the output to outputPath.
// format is the Graphviz output format (e.g. "svg", "png"). If empty, defaults to "svg".
func RenderDOT(dot string, outputPath string, format string) error {
	if format == "" {
		format = "svg"
	}

	// Check that dot binary is available.
	dotBin, err := exec.LookPath("dot")
	if err != nil {
		return fmt.Errorf("graphviz 'dot' command not found: install graphviz (https://graphviz.org/download/) and ensure 'dot' is in your PATH")
	}

	// Write DOT to a temp file.
	tmpFile, err := os.CreateTemp("", "plan-*.dot")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(dot); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing DOT to temp file: %w", err)
	}
	tmpFile.Close()

	// Ensure output directory exists.
	if dir := filepath.Dir(outputPath); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
	}

	// Run: dot -T{format} -o {outputPath} {tmpFile}
	cmd := exec.Command(dotBin, fmt.Sprintf("-T%s", format), "-o", outputPath, tmpFile.Name())
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("dot command failed: %w\n%s", err, string(output))
	}

	return nil
}

// RenderOutputPath determines the output path for a rendered plan diagram.
// It creates a .temp/ directory inside the scope element's directory and returns
// the full path for plan.{format}.
//
// Scope resolution:
//   - scopeID == "" (project level): {boardDir}/.temp/plan.{format}
//   - scopeID is an element: {element.Path}/.temp/plan.{format}
func RenderOutputPath(b *board.Board, scopeID string, name string, format string) (string, error) {
	if format == "" {
		format = "svg"
	}
	if name == "" {
		name = "plan"
	}

	var baseDir string
	if scopeID == "" {
		baseDir = b.Dir
	} else {
		elem := b.FindByID(scopeID)
		if elem == nil {
			return "", fmt.Errorf("element %s not found", scopeID)
		}
		baseDir = elem.Path
	}

	tempDir := filepath.Join(baseDir, ".temp")
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return "", fmt.Errorf("creating .temp directory: %w", err)
	}

	return filepath.Join(tempDir, fmt.Sprintf("%s.%s", name, format)), nil
}
