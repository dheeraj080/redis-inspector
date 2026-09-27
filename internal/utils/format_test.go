package utils

import (
    "testing"
)

func TestFormatBytes(t *testing.T) {
    tests := []struct {
        name     string
        bytes    int64
        expected string
    }{
        {"Zero bytes", 0, "0 B"},
        {"Small bytes", 512, "512 B"},
        {"1 KB boundary", 1024, "1.00 KB"},
        {"1.5 KB", 1536, "1.50 KB"},
        {"1 MB boundary", 1024 * 1024, "1.00 MB"},
        {"2.75 MB", int64(2.75 * 1024 * 1024), "2.75 MB"},
        {"1 GB boundary", 1024 * 1024 * 1024, "1.00 GB"},
        {"3.5 GB", int64(3.5 * 1024 * 1024 * 1024), "3.50 GB"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := FormatBytes(tt.bytes)
            if got != tt.expected {
                t.Errorf("FormatBytes(%d) = %q; want %q", tt.bytes, got, tt.expected)
            }
        })
    }
}

func TestTruncateString(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        maxLen   int
        expected string
    }{
        {"Empty string", "", 10, ""},
        {"Shorter than max", "hello", 10, "hello"},
        {"Equal to max", "hello", 5, "hello"},
        {"Longer than max", "hello world", 8, "hello..."},
        {"MaxLen 3", "hello world", 3, "hel"},
        {"MaxLen 2", "hello world", 2, "he"},
        {"MaxLen 1", "hello world", 1, "h"},
        {"MaxLen 0", "hello world", 0, ""},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := TruncateString(tt.input, tt.maxLen)
            if got != tt.expected {
                t.Errorf("TruncateString(%q, %d) = %q; want %q", tt.input, tt.maxLen, got, tt.expected)
            }
        })
    }
}

func TestRenderSparkline(t *testing.T) {
    t.Run("Empty data returns empty string", func(t *testing.T) {
        if got := RenderSparkline(nil); got != "" {
            t.Errorf("RenderSparkline(nil) = %q; want empty string", got)
        }
        if got := RenderSparkline([]int64{}); got != "" {
            t.Errorf("RenderSparkline([]) = %q; want empty string", got)
        }
    })

    t.Run("Flat data renders uniform bars", func(t *testing.T) {
        data := []int64{100, 100, 100, 100}
        got := RenderSparkline(data)
        runes := []rune(got)
        if len(runes) != len(data) {
            t.Fatalf("length = %d; want %d", len(runes), len(data))
        }
        for i, r := range runes {
            if r != '▂' {
                t.Errorf("rune at %d = %c; want '▂'", i, r)
            }
        }
    })

    t.Run("Varying data renders valid sparkline", func(t *testing.T) {
        data := []int64{10, 20, 30, 40, 50, 60, 70, 80}
        got := RenderSparkline(data)
        runes := []rune(got)
        if len(runes) != len(data) {
            t.Fatalf("length = %d; want %d", len(runes), len(data))
        }
        // First rune should be lowest bar (' '), last rune should be highest bar ('█')
        if runes[0] != ' ' {
            t.Errorf("first rune = %c; want ' '", runes[0])
        }
        if runes[len(runes)-1] != '█' {
            t.Errorf("last rune = %c; want '█'", runes[len(runes)-1])
        }
    })
}
