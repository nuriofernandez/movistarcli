package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
)

var (
	bold      = color.New(color.Bold).SprintFunc()
	colorYes  = color.New(color.FgGreen).SprintFunc()
	colorNo   = color.New(color.FgRed).SprintFunc()
)

// printTable writes header and rows through tabwriter (no ANSI codes), then
// applies bold to the header, colors yes/no in the last column, and runs
// an optional colorRow func on each data line for additional coloring.
func printTable(header string, writeRows func(w *tabwriter.Writer), colorRow func(string) string) {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, header)
	writeRows(w)
	w.Flush()

	lines := strings.Split(buf.String(), "\n")
	for i, line := range lines {
		if i == 0 {
			lines[i] = bold(line)
			continue
		}
		if strings.HasSuffix(line, "yes") {
			line = line[:len(line)-3] + colorYes("yes")
		} else if strings.HasSuffix(line, "no") {
			line = line[:len(line)-2] + colorNo("no")
		}
		if colorRow != nil {
			line = colorRow(line)
		}
		lines[i] = line
	}
	fmt.Print(strings.Join(lines, "\n"))
}
