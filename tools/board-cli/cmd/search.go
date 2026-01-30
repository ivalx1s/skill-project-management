package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/aagrigore/task-board/internal/board"
	"github.com/aagrigore/task-board/internal/output"
	"github.com/spf13/cobra"
)

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
	pattern, err := regexp.Compile("(?i)" + args[0])
	if err != nil {
		return fmt.Errorf("invalid regex: %w", err)
	}

	b, err := board.Load(boardDir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

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
