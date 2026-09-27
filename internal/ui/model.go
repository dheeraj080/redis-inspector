package ui

import (
    "strings"

    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/redis/go-redis/v9"

    rclient "redis-inspector/internal/redis"
)

type Model struct {
    rdb             *redis.Client
    usedMem         string
    peakMem         string
    fragRatio       string
    allocator       string
    topKeys         []rclient.KeyMem
    namespaces      []rclient.NamespaceMem
    viewMode        int
    selected        int
    page            int
    pageSize        int
    searchInput     textinput.Model
    ttlInput        textinput.Model
    isSearching     bool
    confirmDelete   bool
    keyToDelete     string
    showDetails     bool
    activeDetail    rclient.KeyDetail
    showValueViewer bool
    activeValue     string
    showTTLModal    bool
    statusMsg       string
    err             error
    width           int
    height          int
}

func NewModel(rdb *redis.Client) Model {
    ti := textinput.New()
    ti.Placeholder = "Type to filter..."
    ti.CharLimit = 100
    ti.Width = 30

    ttli := textinput.New()
    ttli.Placeholder = "TTL in seconds (-1 to clear)..."
    ttli.CharLimit = 10
    ttli.Width = 25

    return Model{
        rdb:         rdb,
        pageSize:    6,
        searchInput: ti,
        ttlInput:    ttli,
        viewMode:    0,
    }
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(fetchMemoryDataCmd(m.rdb), tickCmd())
}

func (m Model) currentStats() rclient.MemoryStats {
    return rclient.MemoryStats{
        UsedMem:    m.usedMem,
        PeakMem:    m.peakMem,
        FragRatio:  m.fragRatio,
        Allocator:  m.allocator,
        TopKeys:    m.topKeys,
        Namespaces: m.namespaces,
    }
}

func (m Model) filteredKeys() []rclient.KeyMem {
    query := strings.TrimSpace(strings.ToLower(m.searchInput.Value()))
    if query == "" {
        return m.topKeys
    }

    var filtered []rclient.KeyMem
    for _, k := range m.topKeys {
        if strings.Contains(strings.ToLower(k.Key), query) || strings.Contains(strings.ToLower(k.Type), query) {
            filtered = append(filtered, k)
        }
    }
    return filtered
}

func (m Model) filteredNamespaces() []rclient.NamespaceMem {
    query := strings.TrimSpace(strings.ToLower(m.searchInput.Value()))
    if query == "" {
        return m.namespaces
    }

    var filtered []rclient.NamespaceMem
    for _, ns := range m.namespaces {
        if strings.Contains(strings.ToLower(ns.Prefix), query) {
            filtered = append(filtered, ns)
        }
    }
    return filtered
}

func (m Model) totalItems() int {
    if m.viewMode == 1 {
        return len(m.filteredNamespaces())
    }
    return len(m.filteredKeys())
}

func (m Model) maxPages() int {
    total := m.totalItems()
    if total == 0 {
        return 1
    }
    pages := total / m.pageSize
    if total%m.pageSize != 0 {
        pages++
    }
    return pages
}

func (m *Model) clampSelection() {
    total := m.totalItems()
    if total == 0 {
        m.selected = 0
        m.page = 0
        return
    }

    maxP := m.maxPages()
    if m.page >= maxP {
        m.page = maxP - 1
    }
    if m.page < 0 {
        m.page = 0
    }

    start := m.page * m.pageSize
    end := start + m.pageSize
    if end > total {
        end = total
    }

    pageSizeCurrent := end - start
    if m.selected >= pageSizeCurrent {
        m.selected = pageSizeCurrent - 1
    }
    if m.selected < 0 {
        m.selected = 0
    }
}
