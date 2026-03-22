package validate

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// InteractivePrompter prompts the user via terminal for fix decisions
type InteractivePrompter struct {
	reader *bufio.Reader
	writer io.Writer
}

// NewInteractivePrompter creates a prompter that reads from r and writes to w
func NewInteractivePrompter(r io.Reader, w io.Writer) *InteractivePrompter {
	return &InteractivePrompter{
		reader: bufio.NewReader(r),
		writer: w,
	}
}

// Prompt displays a fix suggestion and asks the user what to do
func (p *InteractivePrompter) Prompt(fix Fix, current, total int) FixAction {
	fmt.Fprintf(p.writer, "\n[%d/%d] %s:%d\n", current, total, fix.FilePath, fix.Line)
	fmt.Fprintf(p.writer, "  %s → %s\n", fix.OriginalRaw, fix.CorrectedRaw())
	fmt.Fprintf(p.writer, "  Apply fix? (y)es / (n)o / (a)ll / (q)uit: ")

	line, _ := p.reader.ReadString('\n')
	answer := strings.ToLower(strings.TrimSpace(line))

	switch answer {
	case "y", "yes", "s", "sim":
		return FixActionApply
	case "a", "all", "todos":
		return FixActionApplyAll
	case "q", "quit", "sair":
		return FixActionQuit
	default:
		return FixActionSkip
	}
}

// AutoPrompter automatically applies all fixes without user interaction
type AutoPrompter struct {
	writer io.Writer
}

// NewAutoPrompter creates a prompter that auto-applies all fixes
func NewAutoPrompter(w io.Writer) *AutoPrompter {
	return &AutoPrompter{writer: w}
}

// Prompt displays the fix and auto-applies it
func (p *AutoPrompter) Prompt(fix Fix, current, total int) FixAction {
	fmt.Fprintf(p.writer, "[%d/%d] %s:%d  %s → %s\n", current, total, fix.FilePath, fix.Line, fix.OriginalRaw, fix.CorrectedRaw())
	return FixActionApply
}
