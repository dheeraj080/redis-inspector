package ui

import (
    "fmt"
    "strconv"
    "strings"

    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        if m.showValueViewer {
            switch msg.String() {
            case "esc", "enter", "q", "v":
                m.showValueViewer = false
            }
            return m, nil
        }

        if m.showTTLModal {
            switch msg.String() {
            case "esc":
                m.showTTLModal = false
                m.ttlInput.Blur()
                return m, nil

            case "enter":
                secStr := strings.TrimSpace(m.ttlInput.Value())
                m.showTTLModal = false
                m.ttlInput.Blur()
                m.ttlInput.SetValue("")

                sec, err := strconv.Atoi(secStr)
                if err != nil {
                    m.statusMsg = "Invalid TTL number!"
                    return m, nil
                }
                return m, setTTLCmd(m.rdb, m.activeDetail.Key, sec)

            default:
                m.ttlInput, cmd = m.ttlInput.Update(msg)
                return m, cmd
            }
        }

        if m.showDetails {
            switch msg.String() {
            case "esc", "q":
                m.showDetails = false

            case "v":
                return m, fetchKeyValueCmd(m.rdb, m.activeDetail.Key, m.activeDetail.Type)

            case "t":
                m.showTTLModal = true
                m.ttlInput.Focus()
                return m, textinput.Blink
            }
            return m, nil
        }

        if m.confirmDelete {
            switch msg.String() {
            case "y", "Y":
                m.confirmDelete = false
                return m, deleteKeyCmd(m.rdb, m.keyToDelete)

            case "n", "N", "esc":
                m.confirmDelete = false
                m.keyToDelete = ""

            case "q", "ctrl+c":
                return m, tea.Quit
            }
            return m, nil
        }

        if m.isSearching {
            switch msg.String() {
            case "esc", "enter":
                m.isSearching = false
                m.searchInput.Blur()
                return m, nil
            default:
                m.searchInput, cmd = m.searchInput.Update(msg)
                m.page = 0
                m.clampSelection()
                return m, cmd
            }
        }

        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit

        case "tab":
            m.viewMode = (m.viewMode + 1) % 2
            m.selected = 0
            m.page = 0
            m.clampSelection()

        case "1":
            m.viewMode = 0
            m.selected = 0
            m.page = 0
            m.clampSelection()

        case "2":
            m.viewMode = 1
            m.selected = 0
            m.page = 0
            m.clampSelection()

        case "/":
            m.isSearching = true
            m.statusMsg = ""
            m.searchInput.Focus()
            return m, textinput.Blink

        case "s":
            return m, seedMockDataCmd(m.rdb)

        case "e", "x":
            return m, exportReportCmd(m.currentStats())

        case "v":
            if m.viewMode == 0 {
                fk := m.filteredKeys()
                start := m.page * m.pageSize
                idx := start + m.selected
                if idx >= 0 && idx < len(fk) {
                    return m, fetchKeyValueCmd(m.rdb, fk[idx].Key, fk[idx].Type)
                }
            }

        case "up", "k":
            if m.selected > 0 {
                m.selected--
            } else if m.page > 0 {
                m.page--
                m.selected = m.pageSize - 1
            }

        case "down", "j":
            totalOnPage := m.pageSize
            total := m.totalItems()
            start := m.page * m.pageSize
            end := start + m.pageSize
            if end > total {
                end = total
            }
            if end >= start {
                totalOnPage = end - start
            }

            if m.selected < totalOnPage-1 {
                m.selected++
            } else if m.page < m.maxPages()-1 {
                m.page++
                m.selected = 0
            }

        case "n", "pgdown":
            if m.page < m.maxPages()-1 {
                m.page++
                m.selected = 0
            }

        case "p", "pgup":
            if m.page > 0 {
                m.page--
                m.selected = 0
            }

        case "enter":
            if m.viewMode == 0 {
                fk := m.filteredKeys()
                start := m.page * m.pageSize
                idx := start + m.selected
                if idx >= 0 && idx < len(fk) {
                    return m, fetchKeyDetailsCmd(m.rdb, fk[idx].Key)
                }
            } else if m.viewMode == 1 {
                fn := m.filteredNamespaces()
                start := m.page * m.pageSize
                idx := start + m.selected
                if idx >= 0 && idx < len(fn) {
                    prefix := strings.TrimSuffix(fn[idx].Prefix, "*")
                    if prefix == "(root)" {
                        m.searchInput.SetValue("")
                    } else {
                        m.searchInput.SetValue(prefix)
                    }
                    m.viewMode = 0
                    m.page = 0
                    m.selected = 0
                    m.clampSelection()
                }
            }

        case "d":
            if m.viewMode == 0 {
                fk := m.filteredKeys()
                start := m.page * m.pageSize
                idx := start + m.selected
                if idx >= 0 && idx < len(fk) {
                    m.confirmDelete = true
                    m.keyToDelete = fk[idx].Key
                }
            }
        }

    case KeyDetailMsg:
        if msg.Err != nil {
            m.err = msg.Err
            return m, nil
        }
        m.showDetails = true
        m.activeDetail = msg.Detail

    case KeyValueMsg:
        if msg.Err != nil {
            m.statusMsg = fmt.Sprintf("Error reading value: %v", msg.Err)
            return m, nil
        }
        m.showValueViewer = true
        m.activeValue = msg.Value

    case TTLUpdatedMsg:
        if msg.Err != nil {
            m.statusMsg = fmt.Sprintf("Failed to update TTL: %v", msg.Err)
        } else {
            m.statusMsg = "TTL updated successfully!"
            if m.showDetails {
                return m, fetchKeyDetailsCmd(m.rdb, m.activeDetail.Key)
            }
        }

    case ReportExportedMsg:
        if msg.Err != nil {
            m.statusMsg = fmt.Sprintf("Export failed: %v", msg.Err)
        } else {
            m.statusMsg = fmt.Sprintf("Report saved to %s!", msg.Filename)
        }

    case KeyDeletedMsg:
        m.keyToDelete = ""
        m.statusMsg = "Key unlinked!"
        return m, fetchMemoryDataCmd(m.rdb)

    case KeyDeleteErrMsg:
        m.confirmDelete = false
        m.err = msg.Err

    case DataSeededMsg:
        m.statusMsg = "Mock data seeded!"
        return m, fetchMemoryDataCmd(m.rdb)

    case DataSeedErrMsg:
        m.err = msg.Err

    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        if msg.Height > 15 {
            m.pageSize = msg.Height - 15
        } else {
            m.pageSize = 5
        }
        m.clampSelection()

    case TickMsg:
        return m, tea.Batch(fetchMemoryDataCmd(m.rdb), tickCmd())

    case MemoryDataMsg:
        if msg.Err != nil {
            m.err = msg.Err
            return m, nil
        }
        m.err = nil
        m.usedMem = msg.Stats.UsedMem
        m.peakMem = msg.Stats.PeakMem
        m.fragRatio = msg.Stats.FragRatio
        m.allocator = msg.Stats.Allocator
        m.topKeys = msg.Stats.TopKeys
        m.namespaces = msg.Stats.Namespaces
        m.clampSelection()
    }

    return m, nil
}
