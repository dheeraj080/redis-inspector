package ui

import (
	"testing"

	"github.com/redis/go-redis/v9"
	rclient "github.com/dheeraj080/redis-inspector/internal/redis"
)

func TestModelFiltering(t *testing.T) {
	m := NewModel(nil)
	m.topKeys = []rclient.KeyMem{
		{Key: "user:profile:1", Bytes: 100, Type: "hash"},
		{Key: "user:session:1", Bytes: 200, Type: "string"},
		{Key: "cache:home", Bytes: 50, Type: "string"},
	}
	m.namespaces = []rclient.NamespaceMem{
		{Prefix: "user:profile:*", Bytes: 100, Count: 1},
		{Prefix: "user:session:*", Bytes: 200, Count: 1},
		{Prefix: "cache:*", Bytes: 50, Count: 1},
	}

	t.Run("Empty filter returns all keys", func(t *testing.T) {
		m.searchInput.SetValue("")
		filtered := m.filteredKeys()
		if len(filtered) != 3 {
			t.Fatalf("expected 3 keys, got %d", len(filtered))
		}
	})

	t.Run("Filter by key name", func(t *testing.T) {
		m.searchInput.SetValue("session")
		filtered := m.filteredKeys()
		if len(filtered) != 1 || filtered[0].Key != "user:session:1" {
			t.Fatalf("expected user:session:1, got %+v", filtered)
		}
	})

	t.Run("Filter by type", func(t *testing.T) {
		m.searchInput.SetValue("hash")
		filtered := m.filteredKeys()
		if len(filtered) != 1 || filtered[0].Key != "user:profile:1" {
			t.Fatalf("expected user:profile:1, got %+v", filtered)
		}
	})

	t.Run("Filter namespaces", func(t *testing.T) {
		m.searchInput.SetValue("user")
		filtered := m.filteredNamespaces()
		if len(filtered) != 2 {
			t.Fatalf("expected 2 namespaces, got %d", len(filtered))
		}
	})
}

func TestModelPagination(t *testing.T) {
	m := NewModel(nil)
	m.pageSize = 5
	m.topKeys = make([]rclient.KeyMem, 12)
	for i := 0; i < 12; i++ {
		m.topKeys[i] = rclient.KeyMem{Key: "test", Bytes: int64(i)}
	}

	t.Run("maxPages calculation", func(t *testing.T) {
		if got := m.maxPages(); got != 3 {
			t.Errorf("maxPages() = %d; want 3", got)
		}
	})

	t.Run("clampSelection with out of range page", func(t *testing.T) {
		m.page = 10
		m.selected = 2
		m.clampSelection()
		if m.page != 2 {
			t.Errorf("m.page clamped to %d; want 2", m.page)
		}
		if m.selected != 1 {
			t.Errorf("m.selected clamped to %d; want 1", m.selected)
		}
	})

	t.Run("clampSelection when empty", func(t *testing.T) {
		emptyM := NewModel(nil)
		emptyM.clampSelection()
		if emptyM.page != 0 || emptyM.selected != 0 {
			t.Errorf("empty model clamped to page %d, selected %d; want 0, 0", emptyM.page, emptyM.selected)
		}
	})
}

func TestNewModelOptions(t *testing.T) {
	opts := &redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   4,
	}
	m := NewModel(nil, opts)
	if m.currentDB != 4 {
		t.Errorf("m.currentDB = %d; want 4", m.currentDB)
	}
	if m.redisOpts != opts {
		t.Errorf("m.redisOpts not stored")
	}
}