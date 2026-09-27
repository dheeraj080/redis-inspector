package ui

import (
    "context"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/redis/go-redis/v9"

    rclient "redis-inspector/internal/redis"
)

type KeyValueMsg struct {
    Value string
    Err   error
}

type TTLUpdatedMsg struct {
    Err error
}

type ReportExportedMsg struct {
    Filename string
    Err      error
}

func fetchKeyValueCmd(rdb *redis.Client, key, kType string) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()

        val, err := rclient.FetchKeyValue(ctx, rdb, key, kType)
        return KeyValueMsg{Value: val, Err: err}
    }
}

func setTTLCmd(rdb *redis.Client, key string, seconds int) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()

        err := rclient.SetKeyTTL(ctx, rdb, key, seconds)
        return TTLUpdatedMsg{Err: err}
    }
}

func exportReportCmd(stats rclient.MemoryStats) tea.Cmd {
    return func() tea.Msg {
        filename, err := rclient.ExportReport(stats)
        return ReportExportedMsg{Filename: filename, Err: err}
    }
}

func deleteNamespaceCmd(rdb *redis.Client, pattern string) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
        defer cancel()

        count, err := rclient.DeleteNamespaceKeys(ctx, rdb, pattern)
        return NamespaceDeletedMsg{Count: count, Pattern: pattern, Err: err}
    }
}
