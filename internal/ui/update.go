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
        if m.showEditModal {
            switch msg.String() {
            case "esc":
                m.showEditModal = false
                m.editorInput.Blur()
                return m, nil

            case "enter":
                val := m.editorInput.Value()
                m.showEditModal = false
                m.editorInput.Blur()
                return m, saveKeyValueCmd(m.rdb, m.activeDetail.Key, m.activeDetail.Type, val)

            default:
                m.editorInput, cmd = m.editorInput.Update(msg)
                return m, cmd
            }
        }

        if m.showDBModal {
            switch msg.String() {
            case "esc", "q":
                m.showDBModal = false
                return m, nil

            case "up", "k":
                if m.dbSelected > 0 {
                    m.dbSelected--
                }

            case "down", "j":
                if m.dbSelected < 15 {
                    m.dbSelected++
                }

            case "enter":
                m.showDBModal = false
                return m, switchDBCmd(m.rdb, m.dbSelected)
            }
            return m, nil
        }

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

            case "e":
                m.showEditModal = true
                m.editorInput.SetValue(m.activeValue)
                m.editorInput.Focus()
                return m, textinput.Blink
            }
            return m, nil
        }

        if m.confirmBulkDelete {
            switch msg.String() {
            case "y", "Y":
                m.confirmBulkDelete = false
                return m, deleteNamespaceCmd(m.rdb, m.patternToDelete)

            case "n", "N", "esc":
                m.confirmBulkDelete = false
                m.patternToDelete = ""

            case "q", "ctrl+c":
                return m, tea.Quit
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
            case "esc":
                m.isSearching = false
                m.searchInput.Blur()
                return m, nil

            case "enter":
                m.isSearching = false
                m.searchInput.Blur()
                pattern := m.searchInput.Value()
                if pattern == "" {
                    return m, fetchMemoryDataCmd(m.rdb, m.currentDB)
                }
                m.statusMsg = fmt.Sprintf("Scanning Redis server for pattern: %s...", pattern)
                return m, scanKeysCmd(m.rdb, pattern)

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

        case "b":
            m.showDBModal = true
            m.dbSelected = m.currentDB
            return m, nil

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

        case "x":
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
            } else if m.viewMode == 1 {
                fn := m.filteredNamespaces()
                start := m.page * m.pageSize
                idx := start + m.selected
                if idx >= 0 && idx < len(fn) && fn[idx].Prefix != "(root)" {
                    m.confirmBulkDelete = true
                    m.patternToDelete = fn[idx].Prefix
                }
            }
        }

    case ScannedKeysMsg:
        if msg.Err != nil {
            m.statusMsg = fmt.Sprintf("Scan failed: %v", msg.Err)
        } else {
            m.topKeys = msg.Keys
            m.statusMsg = fmt.Sprintf("Server SCAN found %d keys matching '%s'", len(msg.Keys), msg.Pattern)
            m.page = 0
            m.selected = 0
            m.clampSelection()
        }

    case KeySavedMsg:
        if msg.Err != nil {
            m.statusMsg = fmt.Sprintf("Failed to update key: %v", msg.Err)
        } else {
            m.statusMsg = "Key value updated successfully!"
            if m.showDetails {
                return m, fetchKeyDetailsCmd(m.rdb, m.activeDetail.Key)
            }
        }

    case DBSwitchedMsg:
        if msg.Err != nil {
            m.statusMsg = fmt.Sprintf("Failed to select DB%d: %v", msg.DB, msg.Err)
        } else {
            m.currentDB = msg.DB
            m.statusMsg = fmt.Sprintf("Switched to Database DB%d", msg.DB)
            m.selected = 0
            m.page = 0
            return m, fetchMemoryDataCmd(m.rdb, m.currentDB)
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

    case NamespaceDeletedMsg:
        m.confirmBulkDelete = false
        m.patternToDelete = ""
        if msg.Err != nil {
            m.statusMsg = fmt.Sprintf("Bulk delete failed: %v", msg.Err)
        } else {
            m.statusMsg = fmt.Sprintf("Bulk unlinked %d keys matching '%s'!", msg.Count, msg.Pattern)
        }
        return m, fetchMemoryDataCmd(m.rdb, m.currentDB)

    case KeyDeletedMsg:
        m.keyToDelete = ""
        m.statusMsg = "Key unlinked!"
        return m, fetchMemoryDataCmd(m.rdb, m.currentDB)

    case KeyDeleteErrMsg:
        m.confirmDelete = false
        m.err = msg.Err

    case DataSeededMsg:
        m.statusMsg = "Mock data seeded!"
        return m, fetchMemoryDataCmd(m.rdb, m.currentDB)

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
        return m, tea.Batch(fetchMemoryDataCmd(m.rdb, m.currentDB), tickCmd())

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
        m.databases = msg.Stats.Databases

        if msg.Stats.UsedMemBytes > 0 {
            m.memHistory = append(m.memHistory, msg.Stats.UsedMemBytes)
            if len(m.memHistory) > 30 {
                m.memHistory = m.memHistory[1:]
            }
        }

        m.clampSelection()
    }

    return m, nil
}
