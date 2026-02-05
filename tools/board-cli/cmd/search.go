package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

// SearchResult represents a single search match for JSON output
type SearchResult struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	MatchField   string `json:"matchField"`
	MatchContext string `json:"matchContext"`
}

// SearchResponse is the JSON response for the search command
type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Count   int            `json:"count"`
	Query   string         `json:"query"`
}

var searchCmd = &cobra.Command{
	Use:   "search <regex>",
	Short: "Search board content by regex",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]
	pattern, err := regexp.Compile("(?i)" + query)
	if err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.ValidationError, fmt.Sprintf("invalid regex: %v", err), nil)
			return nil
		}
		return fmt.Errorf("invalid regex: %w", err)
	}

	b, err := board.Load(boardDir)
	if err != nil {
		if JSONEnabled() {
			output.PrintError(os.Stderr, output.InternalError, fmt.Sprintf("loading board: %v", err), nil)
			return nil
		}
		return fmt.Errorf("loading board: %w", err)
	}

	// JSON output mode
	if JSONEnabled() {
		return runSearchJSON(b, pattern, query)
	}

	// Text output mode
	found := 0
	for _, e := range b.Elements {
		files := []string{e.ReadmePath(), e.ProgressPath()}
		for _, f := range files {
			matches := searchFile(f, pattern)
			if len(matches) > 0 {
				ancestry := b.Ancestry(e)
				fmt.Printf("%s%s%s (%s)\n", output.Bold, ancestry, output.Reset, f)
				for _, m := range matches {
					fmt.Printf("  %s%d:%s %s\n", output.Gray, m.line, output.Reset, m.text)
				}
				fmt.Println()
				found += len(matches)
			}
		}
	}

	if found == 0 {
		fmt.Println("No matches found.")
	} else {
		fmt.Printf("%d match(es) found.\n", found)
	}
	return nil
}

func runSearchJSON(b *board.Board, pattern *regexp.Regexp, query string) error {
	// Initialize as empty slice to ensure JSON outputs [] instead of null
	results := []SearchResult{}

	for _, e := range b.Elements {
		files := []string{e.ReadmePath(), e.ProgressPath()}
		for _, f := range files {
			matches := searchFileRaw(f, pattern)
			if len(matches) > 0 {
				// Determine matchField based on file type
				matchField := "readme"
				if filepath.Base(f) == "progress.md" {
					matchField = "progress"
				}

				for _, m := range matches {
					// Highlight matches with **bold** markdown
					highlighted := pattern.ReplaceAllStringFunc(m.text, func(s string) string {
						return "**" + s + "**"
					})

					result := SearchResult{
						ID:           e.ID(),
						Type:         string(e.Type),
						Name:         e.Name,
						Status:       string(e.Status),
						MatchField:   matchField,
						MatchContext: strings.TrimSpace(highlighted),
					}
					results = append(results, result)
				}
			}
		}
	}

	response := SearchResponse{
		Results: results,
		Count:   len(results),
		Query:   query,
	}

	return output.PrintJSON(os.Stdout, response)
}

type searchMatch struct {
	line int
	text string
}

func searchFile(path string, pattern *regexp.Regexp) []searchMatch {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var matches []searchMatch
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		text := scanner.Text()
		if pattern.MatchString(text) {
			highlighted := pattern.ReplaceAllStringFunc(text, func(s string) string {
				return output.Red + s + output.Reset
			})
			matches = append(matches, searchMatch{line: lineNum, text: highlighted})
		}
	}
	return matches
}

// searchFileRaw returns raw matches without ANSI highlighting (for JSON output)
func searchFileRaw(path string, pattern *regexp.Regexp) []searchMatch {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var matches []searchMatch
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		text := scanner.Text()
		if pattern.MatchString(text) {
			matches = append(matches, searchMatch{line: lineNum, text: text})
		}
	}
	return matches
}
