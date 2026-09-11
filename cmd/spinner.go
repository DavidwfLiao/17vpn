package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// spinner draws a single progress line that is redrawn in place and then
// replaced by a final result line.
type spinner struct {
	start time.Time
	frame int
	tty   bool
}

func newSpinner() *spinner {
	return &spinner{
		start: time.Now(),
		tty:   isatty.IsTerminal(os.Stdout.Fd()),
	}
}

// tick redraws the progress line with the next frame, the text and elapsed time.
// It draws nothing when stdout is not a terminal.
func (s *spinner) tick(text string) {
	if !s.tty {
		return
	}
	frame := color.YellowString(spinnerFrames[s.frame%len(spinnerFrames)])
	s.frame++
	fmt.Printf("\r\033[K%s %s %s", frame, text, s.elapsed())
}

// ok replaces the progress line with a green check and the final text.
func (s *spinner) ok(text string) {
	s.finish(color.GreenString("✔"), text)
}

// fail replaces the progress line with a red cross and the final text.
func (s *spinner) fail(text string) {
	s.finish(color.RedString("✖"), text)
}

func (s *spinner) finish(mark, text string) {
	if s.tty {
		fmt.Print("\r\033[K")
	}
	fmt.Printf("%s %s %s\n", mark, text, s.elapsed())
}

func (s *spinner) elapsed() string {
	return color.HiBlackString("%.1fs", time.Since(s.start).Seconds())
}
