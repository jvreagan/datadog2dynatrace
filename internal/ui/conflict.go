package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// PromptConflictResolver returns a ConflictResolver that interactively asks the user
// what to do when a resource already exists in Dynatrace.
//
// Choices per resource:
//
//	r - replace this one
//	s - skip this one
//	R - replace all remaining conflicts
//	S - skip all remaining conflicts
func PromptConflictResolver() dynatrace.ConflictResolver {
	reader := bufio.NewReader(os.Stdin)
	var applyAll *dynatrace.ConflictAction

	return func(resourceType, name string) dynatrace.ConflictAction {
		if applyAll != nil {
			return *applyAll
		}

		label := strings.ReplaceAll(resourceType, "_", " ")
		for {
			fmt.Printf("\n%s %q already exists in Dynatrace.\n", strings.Title(label), name)
			fmt.Print("  [r] replace   [s] skip   [R] replace all   [S] skip all\n> ")

			line, _ := reader.ReadString('\n')
			switch strings.TrimSpace(line) {
			case "r":
				return dynatrace.ConflictReplace
			case "s":
				return dynatrace.ConflictSkip
			case "R":
				a := dynatrace.ConflictReplace
				applyAll = &a
				return dynatrace.ConflictReplace
			case "S":
				a := dynatrace.ConflictSkip
				applyAll = &a
				return dynatrace.ConflictSkip
			default:
				fmt.Println("  Please enter r, s, R, or S.")
			}
		}
	}
}
