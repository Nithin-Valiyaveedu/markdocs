package ui

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var spinnerFrames = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

// Spin shows a spinner with title while fn executes.
// The spinner is cleared when fn returns.
// Returns fn's error.
func Spin(title string, fn func() error) error {
	done := make(chan struct{})
	var fnErr error
	var wg sync.WaitGroup

	// Run the work in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		fnErr = fn()
		close(done)
	}()

	// Animate spinner on the main goroutine
	frame := 0
	spinStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPrimary))
	titleStyle := StyleMuted

	for {
		select {
		case <-done:
			// Clear the spinner line
			fmt.Printf("\r\033[K")
			wg.Wait()
			return fnErr
		default:
			fmt.Printf("\r  %s  %s",
				spinStyle.Render(spinnerFrames[frame%len(spinnerFrames)]),
				titleStyle.Render(title),
			)
			frame++
			time.Sleep(80 * time.Millisecond)
		}
	}
}
