package ui

import (
    "context"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/redis/go-redis/v9"

    rclient "redis-inspector/internal/redis"
)

func tickCmd() tea.Cmd {
    return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
        return TickMsg(t)
    })
}

func fetchMemoryDataCmd(rdb *redis.Client) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        stats, err := rclient.FetchMemoryData(ctx, rdb)
        return MemoryDataMsg{Stats: stats, Err: err}
    }
}

func fetchKeyDetailsCmd(rdb *redis.Client, key string) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        detail, err := rclient.FetchKeyDetails(ctx, rdb, key)
        return KeyDetailMsg{Detail: detail, Err: err}
    }
}

func deleteKeyCmd(rdb *redis.Client, key string) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        if err := rclient.DeleteKey(ctx, rdb, key); err != nil {
            return KeyDeleteErrMsg{Err: err}
        }
        return KeyDeletedMsg{Key: key}
    }
}

func seedMockDataCmd(rdb *redis.Client) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        if err := rclient.SeedMockData(ctx, rdb); err != nil {
            return DataSeedErrMsg{Err: err}
        }
        return DataSeededMsg{}
    }
}
