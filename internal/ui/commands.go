package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/redis/go-redis/v9"
	rclient "github.com/dheeraj080/redis-inspector/internal/redis"
)

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func fetchMemoryDataCmd(rdb *redis.Client, db int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		stats, err := rclient.FetchMemoryData(ctx, rdb, db)
		return MemoryDataMsg{Stats: stats, Err: err}
	}
}

func fetchKeyDetailsCmd(rdb *redis.Client, key string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		detail, err := rclient.FetchKeyDetails(ctx, rdb, key)
		return KeyDetailMsg{Detail: detail, Err: err}
	}
}

func scanKeysCmd(rdb *redis.Client, pattern string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		keys, err := rclient.FetchKeysByPattern(ctx, rdb, pattern)
		return ScannedKeysMsg{Keys: keys, Pattern: pattern, Err: err}
	}
}

func saveKeyValueCmd(rdb *redis.Client, key, kType, newValue string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err := rclient.SaveKeyValue(ctx, rdb, key, kType, newValue)
		return KeySavedMsg{Err: err}
	}
}

func deleteKeyCmd(rdb *redis.Client, key string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err := rclient.DeleteKey(ctx, rdb, key)
		if err != nil {
			return KeyDeleteErrMsg{Err: err}
		}
		return KeyDeletedMsg{Key: key}
	}
}

func seedMockDataCmd(rdb *redis.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := rclient.SeedMockData(ctx, rdb)
		if err != nil {
			return DataSeedErrMsg{Err: err}
		}
		return DataSeededMsg{}
	}
}

func switchDBCmd(opts *redis.Options, db int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		newClient, err := rclient.NewClientWithDB(ctx, opts, db)
		if err != nil {
			return DBSwitchedMsg{DB: db, Client: nil, Err: err}
		}
		return DBSwitchedMsg{DB: db, Client: newClient, Err: nil}
	}
}
