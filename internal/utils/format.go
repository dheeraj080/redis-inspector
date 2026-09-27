package utils

import "fmt"

func FormatBytes(bytes int64) string {
    const (
        _  = iota
        KB = 1 << (10 * iota)
        MB
        GB
        TB
    )

    switch {
    case bytes >= GB:
        return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
    case bytes >= MB:
        return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
    case bytes >= KB:
        return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
    default:
        return fmt.Sprintf("%d B", bytes)
    }
}

func TruncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    if maxLen <= 3 {
        return s[:maxLen]
    }
    return s[:maxLen-3] + "..."
}

func RenderSparkline(data []int64) string {
    if len(data) == 0 {
        return ""
    }

    bars := []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

    min, max := data[0], data[0]
    for _, v := range data {
        if v < min {
            min = v
        }
        if v > max {
            max = v
        }
    }

    delta := max - min
    res := make([]rune, len(data))

    for i, v := range data {
        if delta == 0 {
            res[i] = bars[1]
        } else {
            idx := int(float64(v-min) / float64(delta) * float64(len(bars)-1))
            if idx < 0 {
                idx = 0
            }
            if idx >= len(bars) {
                idx = len(bars) - 1
            }
            res[i] = bars[idx]
        }
    }

    return string(res)
}
