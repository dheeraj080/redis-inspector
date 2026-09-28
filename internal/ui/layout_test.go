package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestResponsiveLayoutSizing(t *testing.T) {
	tests := []struct {
		width int
		want  int
	}{
		{20, 1},
		{40, 2},
		{69, 2},
		{76, 4},
	}
	for _, tt := range tests {
		if got := metricColumnCount(maxInt(1, tt.width-6)); got != tt.want {
			t.Errorf("metricColumnCount(%d) = %d; want %d", tt.width-6, got, tt.want)
		}
	}
	
	if got := tableKeyColumnWidth(34); got > 22 {
		t.Errorf("narrow table key width = %d; expected it to stay compact", got)
	}
	
	// Updated to expect 38 (80 - 42) to match the conservative width calculation
	// that prevents right-edge truncation in the UI.
	if got := tableKeyColumnWidth(80); got != 38 {
		t.Errorf("wide table key width = %d; want 38", got)
	}
	
	if got := lipgloss.Width(renderHeader(30, "brand", "db", "status", "localhost:6379")); got > 30 {
		t.Errorf("renderHeader exceeded width: %d", got)
	}
	
	if got := maxLineWidth(fitTerminalWidth("1234567890\nabcdefghijklmno", 10)); got > 10 {
		t.Errorf("fitTerminalWidth exceeded width: %d", got)
	}
}

func maxLineWidth(s string) int {
	maxWidth := 0
	for _, line := range strings.Split(s, "\n") {
		if width := lipgloss.Width(line); width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth
}